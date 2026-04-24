use std::{collections::HashMap, sync::Arc};

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::api_error::ApiError;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Paste {
    pub id: String,
    pub title: String,
    pub content: String,
    pub language: String,
    pub created_at: DateTime<Utc>,
    pub expires_at: Option<DateTime<Utc>>,
    #[serde(skip_serializing)]
    pub password_hash: String,
    pub protected: bool,
}

#[derive(Clone)]
pub struct Store {
    pastes: Arc<RwLock<HashMap<String, Paste>>>,
}

impl Store {
    pub fn new() -> Self {
        Store {
            pastes: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn create(&self, mut paste: Paste) -> Result<Paste, ApiError> {
        paste.id = Uuid::new_v4().to_string();
        paste.created_at = Utc::now();
        let mut pastes = self.pastes.write().await;
        pastes.insert(paste.id.clone(), paste.clone());
        Ok(paste)
    }

    pub async fn get(&self, id: Uuid) -> Result<Paste, ApiError> {
        let pastes = self.pastes.read().await;
        pastes
            .get(&id.to_string())
            .cloned()
            .ok_or(ApiError::NotFound)
    }

    pub async fn delete(&self, id: Uuid) -> Result<(), ApiError> {
        let mut pastes = self.pastes.write().await;
        pastes
            .remove(&id.to_string())
            .map(|_| ())
            .ok_or(ApiError::NotFound)
    }

    pub async fn remove_expired(&self) {
        let mut pastes = self.pastes.write().await;
        pastes.retain(|_, p| p.expires_at.map_or(true, |exp| exp > Utc::now()));
    }
}

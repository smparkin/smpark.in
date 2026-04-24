use axum::{
    Json,
    extract::{Path, State},
    http::{HeaderMap, StatusCode},
    response::IntoResponse,
};
use chrono::{Duration, Utc};
use serde::Deserialize;
use serde_json::{Value, json};
use uuid::Uuid;

use crate::{
    api_error::ApiError,
    paste::store::{Paste, Store},
};

#[derive(Deserialize)]
pub struct CreateRequest {
    title: String,
    content: String,
    language: String,
    expiry: i64,
    password: Option<String>,
}

pub(crate) async fn create_paste(
    State(store): State<Store>,
    Json(req): Json<CreateRequest>,
) -> Result<impl IntoResponse, ApiError> {
    if req.content.is_empty() {
        return Err(ApiError::BadRequest("content is required".into()));
    }
    if req.content.len() > 1_000_000 {
        return Err(ApiError::BadRequest(
            "content exceeds 1,000,000 characters".into(),
        ));
    }
    if req.title.len() > 200 {
        return Err(ApiError::BadRequest("title exceeds 200 characters".into()));
    }
    if req.expiry < 1 || req.expiry > 10 {
        return Err(ApiError::BadRequest(
            "expiry must be between 1 and 10 minutes".into(),
        ));
    }

    let (password_hash, protected) = if let Some(pw) = req.password {
        if pw.len() > 72 {
            return Err(ApiError::BadRequest(
                "password exceeds 72 characters".into(),
            ));
        }
        let hash = bcrypt::hash(pw, bcrypt::DEFAULT_COST).map_err(|_| ApiError::InternalError)?;
        (hash, true)
    } else {
        (String::new(), false)
    };

    let paste = Paste {
        id: String::new(),
        title: req.title,
        content: req.content,
        language: req.language,
        created_at: Utc::now(),
        expires_at: Some(Utc::now() + Duration::minutes(req.expiry)),
        password_hash,
        protected,
    };

    let paste = store.create(paste).await?;

    Ok((
        StatusCode::CREATED,
        Json(json!({
            "id": paste.id,
            "url": format!("/paste/{}", paste.id),
        })),
    ))
}

pub(crate) async fn get_paste(
    State(store): State<Store>,
    Path(id): Path<String>,
    headers: HeaderMap,
) -> Result<Json<Value>, ApiError> {
    let id = Uuid::parse_str(&id).map_err(|_| ApiError::BadRequest("invalid id".into()))?;
    let paste = store.get(id).await?;

    if paste.expires_at.map_or(false, |exp| Utc::now() > exp) {
        store.delete(id).await.ok();
        return Err(ApiError::Gone);
    }

    if paste.protected {
        let pw = headers
            .get("X-Paste-Password")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");
        if bcrypt::verify(pw, &paste.password_hash).unwrap_or(false) == false {
            return Err(ApiError::Forbidden);
        }
    }

    Ok(Json(
        serde_json::to_value(&paste).map_err(|_| ApiError::InternalError)?,
    ))
}

pub(crate) async fn get_raw_paste(
    State(store): State<Store>,
    Path(id): Path<String>,
    headers: HeaderMap,
) -> Result<impl IntoResponse, ApiError> {
    let id = Uuid::parse_str(&id).map_err(|_| ApiError::BadRequest("invalid id".into()))?;
    let paste = store.get(id).await?;

    if paste.expires_at.map_or(false, |exp| Utc::now() > exp) {
        store.delete(id).await.ok();
        return Err(ApiError::Gone);
    }

    if paste.protected {
        let pw = headers
            .get("X-Paste-Password")
            .and_then(|v| v.to_str().ok())
            .unwrap_or("");
        if bcrypt::verify(pw, &paste.password_hash).unwrap_or(false) == false {
            return Err(ApiError::Forbidden);
        }
    }

    Ok((
        StatusCode::OK,
        [
            ("Content-Type", "text/plain; charset=utf-8"),
            ("X-Content-Type-Options", "nosniff"),
        ],
        paste.content,
    ))
}

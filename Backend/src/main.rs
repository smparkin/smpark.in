use axum::{
    Json, Router,
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
};
use serde_json::json;
mod api_error;
use api_error::ApiError;
use tower_http::services::{ServeDir, ServeFile};

use crate::paste::store::Store;
mod paste;

impl IntoResponse for ApiError {
    fn into_response(self) -> axum::response::Response {
        let (status, error_message) = match self {
            ApiError::NotFound => (StatusCode::NOT_FOUND, "not found".to_string()),
            ApiError::BadRequest(msg) => (StatusCode::BAD_REQUEST, msg),
            ApiError::InternalError => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "failed to process request".to_string(),
            ),
            ApiError::Gone => (StatusCode::GONE, "paste expired".to_string()),
            ApiError::Forbidden => (StatusCode::FORBIDDEN, "password required".to_string()),
        };

        let body = Json(json!({
            "error": error_message,
        }));

        (status, body).into_response()
    }
}

fn create_app(store: Store) -> Router {
    Router::new()
        .route("/api/paste", post(paste::handlers::create_paste))
        .route("/api/paste/{id}", get(paste::handlers::get_paste))
        .route("/api/paste/{id}/raw", get(paste::handlers::get_raw_paste))
        .fallback_service(
            ServeDir::new("./public").fallback(ServeFile::new("./public/app/index.html")),
        )
        .with_state(store)
}

#[tokio::main]
async fn main() {
    let store = Store::new();
    let expiry_store = store.clone();

    tokio::spawn(async move {
        loop {
            tokio::time::sleep(tokio::time::Duration::from_secs(30)).await;
            expiry_store.remove_expired().await;
        }
    });

    let app = create_app(store);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:8080")
        .await
        .expect("failed to bind to port 8080");

    println!("Listening on http://localhost:8080");

    axum::serve(listener, app)
        .await
        .expect("failed to start axum")
}

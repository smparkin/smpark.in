#[derive(Debug)]
pub(crate) enum ApiError {
    NotFound,
    BadRequest(String),
    InternalError,
    Gone,
    Forbidden,
}

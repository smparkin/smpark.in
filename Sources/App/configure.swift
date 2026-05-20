import Leaf
import Vapor

public func configure(_ app: Application) throws {
    app.views.use(.leaf)

    app.middleware.use(FileMiddleware(publicDirectory: app.directory.publicDirectory))
    app.middleware.use(ErrorMiddleware.default(environment: app.environment))

    let encoder = JSONEncoder()
    encoder.dateEncodingStrategy = .iso8601
    ContentConfiguration.global.use(encoder: encoder, for: .json)

    let store = PasteStore()
    app.pasteStore = store
    Task { await store.startExpiryWorker() }

    try routes(app)
}

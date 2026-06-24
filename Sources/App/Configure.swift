import Vapor

public func configure(_ app: Application) throws {
    app.middleware.use(FileMiddleware(publicDirectory: app.directory.publicDirectory, defaultFile: "index.html"))
    app.middleware.use(ErrorMiddleware.default(environment: app.environment))

    try routes(app)
}

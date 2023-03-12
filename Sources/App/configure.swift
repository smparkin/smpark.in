import Leaf
import Vapor

// Called before your application initializes.
public func configure(_ app: Application) throws {
    //leaf
    app.views.use(.leaf)
    
    //Register middleware
    app.middleware.use(FileMiddleware(publicDirectory: app.directory.publicDirectory))
    app.middleware.use(ErrorMiddleware.default(environment: app.environment))
    
    //routes
    try routes(app)
}

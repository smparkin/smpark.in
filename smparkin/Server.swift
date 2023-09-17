import Leaf
import Vapor
import SwiftUI

class Server: ObservableObject {
    @Published var logs: [String] = [String]()
    let dateFormatter = ISO8601DateFormatter()
    var app: Application
    let port: Int
    
    init(port: Int) {
        self.port = port
        app = Application(.production)
        configure(app)
    }
    
    // Called before your application initializes.
    private func configure(_ app: Application) {
        app.http.server.configuration.hostname = "0.0.0.0"
        app.http.server.configuration.port = port
        
        //leaf
        app.views.use(.leaf)
        app.leaf.cache.isEnabled = app.environment.isRelease
        app.leaf.configuration.rootDirectory = Bundle.main.bundlePath
        
        //Register middleware
        app.middleware.use(FileMiddleware(publicDirectory: "\(Bundle.main.bundlePath)/\(app.directory.publicDirectory)"))
        app.middleware.use(ErrorMiddleware.default(environment: app.environment))
        app.middleware.use(LogMiddleware.init(server: self))
        
        //routes
        do {
            try routes(app)
        } catch {
            fatalError(error.localizedDescription)
        }
    }
    
    func start() {
        Task(priority: .background) {
            do {
                try app.start()
            } catch {
                fatalError(error.localizedDescription)
            }
        }
    }
    
    func stop() {
        app.shutdown()
    }
    
    func restart() {
        stop()
        start()
    }
}

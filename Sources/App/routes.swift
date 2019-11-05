import Vapor

// Register your application's routes here.
public func routes(_ router: Router) throws {
    router.get("/") { request in
        return try request.view().render("home")
    }
    router.get("/index.html") { request in
        return try request.view().render("home")
    }
    router.get("/privacy") { request in
        return try request.view().render("privacy")
    }
}

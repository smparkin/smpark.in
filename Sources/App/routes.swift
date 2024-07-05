import Vapor

// Register your application's routes here.
public func routes(_ app: Application) throws {
    app.get("") { req async throws -> View in
        return try await req.view.render("home")
    }
    app.get("index.html") { req async throws -> View in
        return try await req.view.render("home")
    }
    app.get("privacy") { req async throws -> View in
        return try await req.view.render("privacy")
    }
    app.get("welcome") { req async throws -> View in
        return try await req.view.render("welcome")
    }
    app.get("**") { req async throws -> View in
        return try await req.view.render("404")
    }
}

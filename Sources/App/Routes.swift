import Vapor

private func currentYear() -> Int {
    Calendar.current.component(.year, from: Date())
}

public func routes(_ app: Application) throws {
    app.get("") { req async throws -> View in
        try await req.view.render("home", ["year": currentYear()])
    }
    app.get("index.html") { req async throws -> View in
        try await req.view.render("home", ["year": currentYear()])
    }
    app.get("resume") { req async throws -> View in
        try await req.view.render("resume", ["year": currentYear()])
    }
    app.get("paste") { req async throws -> View in
        try await req.view.render("paste")
    }
    app.get("paste", ":id") { req async throws -> View in
        let id = req.parameters.get("id") ?? ""
        return try await req.view.render("paste-view", ["pasteID": id])
    }

    let rl = RateLimitMiddleware()
    app.grouped(rl).post("api", "paste", use: pasteCreateHandler)
    app.get("api", "paste", ":id", use: pasteGetHandler)
    app.get("api", "paste", ":id", "raw", use: pasteGetRawHandler)

    app.get("**") { req async throws -> View in
        try await req.view.render("404")
    }
}

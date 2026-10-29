import Foundation
import Vapor

struct PasteCreateRequest: Content {
    let title: String?
    let content: String
    let language: String?
    let expiry: Int
    let password: String?
}

struct PasteCreateResponse: Content {
    let id: String
    let url: String
}

struct PastePublic: Content {
    let id: String
    let title: String
    let content: String
    let language: String
    let createdAt: Date
    let expiresAt: Date?
    let isProtected: Bool

    enum CodingKeys: String, CodingKey {
        case id, title, content, language
        case createdAt = "created_at"
        case expiresAt = "expires_at"
        case isProtected = "protected"
    }
}

struct ErrorBody: Content {
    let error: String
}

struct LockedBody: Content {
    let protected: Bool
    let error: String
}

func pasteCreateHandler(_ req: Request) async throws -> Response {
    let body = try req.content.decode(PasteCreateRequest.self)

    guard !body.content.isEmpty else {
        return errorResponse(req, .badRequest, "content is required")
    }
    guard body.content.count <= 1_000_000 else {
        return errorResponse(req, .payloadTooLarge, "content exceeds 1,000,000 characters")
    }
    guard (body.title ?? "").count <= 200 else {
        return errorResponse(req, .badRequest, "title exceeds 200 characters")
    }
    guard (1...10).contains(body.expiry) else {
        return errorResponse(req, .badRequest, "expiry must be between 1 and 10 minutes")
    }

    var hash: String?
    if let pw = body.password, !pw.isEmpty {
        guard pw.count <= 72 else {
            return errorResponse(req, .badRequest, "password exceeds 72 characters")
        }
        hash = try Bcrypt.hash(pw)
    }

    let paste = Paste(
        id: "",
        title: body.title ?? "",
        content: body.content,
        language: body.language ?? "plaintext",
        createdAt: Date(),
        expiresAt: Date().addingTimeInterval(Double(body.expiry) * 60),
        passwordHash: hash,
        isProtected: hash != nil
    )

    let store = req.application.pasteStore
    let id = await store.create(paste)

    let resp = Response(status: .created)
    try resp.content.encode(PasteCreateResponse(id: id, url: "/paste/\(id)"))
    return resp
}

func pasteGetHandler(_ req: Request) async throws -> Response {
    guard let id = req.parameters.get("id") else {
        return errorResponse(req, .badRequest, "missing id")
    }

    let store = req.application.pasteStore
    guard let paste = await store.get(id) else {
        return errorResponse(req, .notFound, "paste not found")
    }

    if let exp = paste.expiresAt, Date() > exp {
        await store.delete(id)
        return errorResponse(req, .gone, "paste has expired")
    }

    if paste.isProtected {
        let pw = req.headers.first(name: "X-Paste-Password")
        let valid = pw.map {
            !$0.isEmpty && (try? Bcrypt.verify($0, created: paste.passwordHash ?? "")) == true
        } ?? false
        if !valid {
            let resp = Response(status: .forbidden)
            try resp.content.encode(LockedBody(protected: true, error: "password required"))
            return resp
        }
    }

    let pub = PastePublic(
        id: paste.id,
        title: paste.title,
        content: paste.content,
        language: paste.language,
        createdAt: paste.createdAt,
        expiresAt: paste.expiresAt,
        isProtected: paste.isProtected
    )
    let resp = Response(status: .ok)
    try resp.content.encode(pub)
    return resp
}

func pasteGetRawHandler(_ req: Request) async throws -> Response {
    guard let id = req.parameters.get("id") else {
        return Response(status: .badRequest, body: .init(string: "missing id"))
    }

    let store = req.application.pasteStore
    guard let paste = await store.get(id) else {
        return Response(status: .notFound, body: .init(string: "paste not found"))
    }

    if let exp = paste.expiresAt, Date() > exp {
        await store.delete(id)
        return Response(status: .gone, body: .init(string: "paste has expired"))
    }

    if paste.isProtected {
        let pw = req.headers.first(name: "X-Paste-Password")
        let valid = pw.map {
            !$0.isEmpty && (try? Bcrypt.verify($0, created: paste.passwordHash ?? "")) == true
        } ?? false
        if !valid {
            return Response(status: .forbidden, body: .init(string: "password required"))
        }
    }

    var headers = HTTPHeaders()
    headers.contentType = .plainText
    headers.add(name: "X-Content-Type-Options", value: "nosniff")
    return Response(status: .ok, headers: headers, body: .init(string: paste.content))
}

private func errorResponse(_ req: Request, _ status: HTTPStatus, _ message: String) -> Response {
    let resp = Response(status: status)
    try? resp.content.encode(ErrorBody(error: message))
    return resp
}

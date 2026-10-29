import Vapor

struct RateLimitMiddleware: AsyncMiddleware {
    let state = RateLimiterState()

    func respond(to request: Request, chainingTo next: AsyncResponder) async throws -> Response {
        let ip = extractIP(from: request)
        guard await state.allow(ip: ip) else {
            request.logger.warning("rate limit exceeded", metadata: ["ip": .string(ip)])
            let resp = Response(status: .tooManyRequests)
            try? resp.content.encode(["error": "rate limit exceeded"])
            return resp
        }
        return try await next.respond(to: request)
    }

    private func extractIP(from req: Request) -> String {
        if let fwd = req.headers.first(name: .xForwardedFor) {
            return fwd.split(separator: ",").first.map(String.init)?.trimmingCharacters(in: .whitespaces) ?? fwd
        }
        return req.remoteAddress?.ipAddress ?? "unknown"
    }
}

actor RateLimiterState {
    private struct Entry {
        var count: Int
        var resetAt: Date
    }
    private var clients: [String: Entry] = [:]

    func allow(ip: String) -> Bool {
        let now = Date()
        if let entry = clients[ip] {
            if now < entry.resetAt {
                if entry.count >= 1 { return false }
                clients[ip] = Entry(count: entry.count + 1, resetAt: entry.resetAt)
            } else {
                clients[ip] = Entry(count: 1, resetAt: now.addingTimeInterval(30))
            }
        } else {
            clients[ip] = Entry(count: 1, resetAt: now.addingTimeInterval(30))
        }
        return true
    }
}

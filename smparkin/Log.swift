//
//  Log.swift
//  smparkin
//
//  Created by Stephen Parkinson on 9/11/23.
//

import Vapor

struct LogMiddleware: Middleware {
    let server: Server
    
    init(server: Server) {
        self.server = server
    }
    
    func respond(to request: Request, chainingTo next: Responder) -> EventLoopFuture<Response> {
        DispatchQueue.main.async {
            server.logs.append("\(server.dateFormatter.string(from: Date.now)) \(request.url.path)")
        }
        return next.respond(to: request)
    }
}

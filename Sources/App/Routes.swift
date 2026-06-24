import Vapor
import Foundation

public func routes(_ app: Application) throws {
    app.get { req in try serveThemePage(req, app: app, file: "Home.html") }
    app.get("resume") { req in try serveThemePage(req, app: app, file: "Resume.html") }
    app.get("feed", ":theme", "feed.xml") { req in try serveThemeFeed(req, app: app) }
}

/// Names of the theme directories under Public/themes.
private func availableThemes(_ app: Application) -> [String] {
    let themesPath = app.directory.publicDirectory + "themes"
    let entries = (try? FileManager.default.contentsOfDirectory(atPath: themesPath)) ?? []
    return entries.filter { name in
        guard !name.hasPrefix(".") else { return false }
        var isDir: ObjCBool = false
        FileManager.default.fileExists(atPath: "\(themesPath)/\(name)", isDirectory: &isDir)
        return isDir.boolValue
    }
}

/// Serves a theme page (picked at random) with its asset paths and navbar rewritten
/// so the page renders correctly from the clean "/" and "/resume" URLs.
private func serveThemePage(_ req: Request, app: Application, file: String) throws -> Response {
    guard let theme = availableThemes(app).randomElement() else {
        throw Abort(.internalServerError, reason: "No themes found")
    }
    let htmlPath = app.directory.publicDirectory + "themes/\(theme)/\(file)"
    guard var html = try? String(contentsOfFile: htmlPath, encoding: .utf8) else {
        throw Abort(.internalServerError, reason: "Could not read \(file) for theme \(theme)")
    }

    // A <base> tag rewrites all relative asset paths (including ones built at
    // runtime by iWebImage.js) to the theme directory.
    html = html.replacingOccurrences(of: "<head>", with: "<head><base href=\"/themes/\(theme)/\">")

    // The navbar prepends "path-to-root" to each feed link. Setting it to "/" makes
    // the Home/Resume links absolute ("/" and "/resume") so they bypass the <base>
    // and hit our routes instead of the theme directory.
    html = html.replacingOccurrences(of: "\"path-to-root\": \"\"", with: "\"path-to-root\": \"/\"")

    // Treat each page as its own collection so the navbar keeps a real href on the
    // current-page link instead of blanking it (an empty href would resolve against
    // the <base> to the theme dir, whose index.html redirects to Home.html). The
    // current-page highlight styling is unaffected.
    html = html.replacingOccurrences(of: "\"isCollectionPage\": \"NO\"", with: "\"isCollectionPage\": \"YES\"")

    // Point the navbar's feed fetch at our /feed/<theme> route so we can rewrite the
    // page links on the fly (the static feed.xml would otherwise be served as-is).
    html = html.replacingOccurrences(
        of: "'Scripts/Widgets/SharedResources', '.'",
        with: "'Scripts/Widgets/SharedResources', '/feed/\(theme)'"
    )

    return Response(status: .ok, headers: ["Content-Type": "text/html; charset=utf-8"], body: .init(string: html))
}

/// Serves a theme's navbar feed with the page links pointed at our clean URLs.
private func serveThemeFeed(_ req: Request, app: Application) throws -> Response {
    let theme = req.parameters.get("theme") ?? ""
    guard availableThemes(app).contains(theme) else {
        throw Abort(.notFound)
    }
    let feedPath = app.directory.publicDirectory + "themes/\(theme)/feed.xml"
    guard var feed = try? String(contentsOfFile: feedPath, encoding: .utf8) else {
        throw Abort(.notFound)
    }

    // Repoint the navbar entries. Combined with the page's path-to-root of "/",
    // "." becomes "/." (normalizes to "/") and "resume" becomes "/resume". The home
    // href is kept non-empty to dodge a typo bug in navbar.js that throws on empty
    // hrefs and would leave the whole navbar empty.
    feed = feed
        .replacingOccurrences(of: "<link rel=\"alternate\" href=\"Home.html\"/>", with: "<link rel=\"alternate\" href=\".\"/>")
        .replacingOccurrences(of: "<link rel=\"alternate\" href=\"Resume.html\"/>", with: "<link rel=\"alternate\" href=\"resume\"/>")

    return Response(status: .ok, headers: ["Content-Type": "application/atom+xml; charset=utf-8"], body: .init(string: feed))
}

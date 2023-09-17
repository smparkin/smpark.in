//
//  ContentView.swift
//  smparkin
//
//  Created by Stephen Parkinson on 9/11/23.
//

import SwiftUI

struct ContentView: View {
    @StateObject var server: Server = Server(port: 8080)
    
    var body: some View {
        VStack {
            Text(ProcessInfo().hostName + ":\(server.port)")
            List(server.logs.reversed(), id: \.self) { log in
                Text(log)
            }
        }
        .padding()
        .onAppear {
            server.start()
        }
    }
}

#Preview {
    ContentView()
}

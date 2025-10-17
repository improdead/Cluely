// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "Cluely",
    platforms: [
        .visionOS(.v1)
    ],
    products: [
        .executable(name: "Cluely", targets: ["Cluely"])
    ],
    targets: [
        .executableTarget(
            name: "Cluely",
            path: "CluelyApp",
            resources: [
                .process("Resources")
            ],
            swiftSettings: [
                .enableUpcomingFeature("BareSlashRegexLiterals")
            ],
            linkerSettings: [
                .linkedFramework("AVFoundation"),
                .linkedFramework("Vision"),
                .linkedFramework("ReplayKit"),
                .linkedFramework("Accelerate"),
                .linkedFramework("SwiftUI")
            ]
        )
    ]
)

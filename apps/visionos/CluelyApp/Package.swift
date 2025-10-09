// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "CluelyApp",
    platforms: [
        .visionOS(.v1)
    ],
    products: [
        .library(
            name: "CluelyApp",
            targets: ["CluelyApp"]
        )
    ],
    targets: [
        .target(
            name: "CluelyApp",
            dependencies: [],
            path: ".",
            sources: [
                "App.swift",
                "Config.swift",
                "Input/VAD.swift",
                "Session/AudioEngine.swift",
                "Session/CoachRuntime.swift",
                "Session/FrameSource.swift",
                "Session/Models.swift",
                "Session/SessionController.swift",
                "Session/WSClient.swift",
                "Views/CoachChip.swift",
                "Views/CoachView.swift",
                "Views/ContextCard.swift",
                "Views/FramePreviewCard.swift",
                "Views/ListeningIndicator.swift",
                "Views/OrnamentControls.swift",
                "Views/RootWindow.swift",
                "Views/SuggestionPanel.swift",
                "Views/WarningBadge.swift"
            ],
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

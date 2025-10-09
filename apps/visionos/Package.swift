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
    dependencies: [
        // Add any dependencies here if needed
    ],
    targets: [
        .executableTarget(
            name: "Cluely",
            resources: [
                .copy("CluelyApp/Resources/Info.plist")
            ]
        )
    ]
)

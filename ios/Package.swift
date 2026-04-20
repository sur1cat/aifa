// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "HabitFlow",
    platforms: [
        .iOS(.v17)
    ],
    products: [
        .library(
            name: "HabitFlow",
            targets: ["HabitFlow"]
        ),
    ],
    dependencies: [
        .package(url: "https://github.com/google/GoogleSignIn-iOS.git", from: "7.0.0"),
    ],
    targets: [
        .target(
            name: "HabitFlow",
            dependencies: [
                .product(name: "GoogleSignIn", package: "GoogleSignIn-iOS"),
            ],
            path: "HabitFlow"
        ),
    ]
)

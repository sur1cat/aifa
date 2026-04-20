# iOS App Setup

## Prerequisites

- Xcode 15.0+
- iOS 17.0+ target
- Google Cloud Console account (for Google Sign-In)
- Apple Developer account (for Apple Sign-In)

## 1. Create Xcode Project

1. Open Xcode
2. File > New > Project
3. Select "App" template
4. Configure:
   - Product Name: `HabitFlow`
   - Team: Your team
   - Organization Identifier: `com.yourname`
   - Interface: SwiftUI
   - Language: Swift
   - Storage: SwiftData
5. Save to `/habitflow/ios/` folder

## 2. Add Dependencies

1. In Xcode: File > Add Package Dependencies
2. Add Google Sign-In:
   - URL: `https://github.com/google/GoogleSignIn-iOS.git`
   - Version: 7.0.0 or later

## 3. Configure Google Sign-In

### Get Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Go to APIs & Services > Credentials
4. Create OAuth 2.0 Client ID:
   - Application type: iOS
   - Bundle ID: `com.yourname.HabitFlow`
5. Download the configuration file

### Configure in Xcode

1. Open `Info.plist`
2. Add URL Scheme:
   ```xml
   <key>CFBundleURLTypes</key>
   <array>
       <dict>
           <key>CFBundleURLSchemes</key>
           <array>
               <string>com.googleusercontent.apps.YOUR_CLIENT_ID</string>
           </array>
       </dict>
   </array>
   ```

3. Add Client ID:
   ```xml
   <key>GIDClientID</key>
   <string>YOUR_CLIENT_ID.apps.googleusercontent.com</string>
   ```

## 4. Configure Apple Sign-In

1. In Xcode: Target > Signing & Capabilities
2. Add "Sign in with Apple" capability
3. Ensure your Apple Developer account has this enabled

## 5. Copy Source Files

Copy the following files to your Xcode project:

```
HabitFlow/
├── App/
│   └── HabitFlowApp.swift
├── Core/
│   ├── Auth/
│   │   ├── AuthManager.swift
│   │   └── AuthModels.swift
│   ├── Network/
│   │   └── APIClient.swift
│   ├── Storage/
│   │   └── KeychainHelper.swift
│   └── DesignSystem/
│       ├── Colors.swift
│       ├── Typography.swift
│       └── Spacing.swift
└── Features/
    └── Auth/
        └── LoginView.swift
```

## 6. Update API URL

In `APIClient.swift`, update the base URL:

```swift
#if DEBUG
self.baseURL = URL(string: "http://localhost:8080/api/v1")!
#else
self.baseURL = URL(string: "https://your-api.com/api/v1")!
#endif
```

## 7. Run the App

1. Start the backend:
   ```bash
   cd backend
   ./bin/api
   ```

2. Run in Xcode:
   - Select your target device/simulator
   - Press Cmd+R

## Troubleshooting

### Google Sign-In not working

- Verify URL scheme matches reversed client ID
- Check bundle ID matches Google Cloud configuration
- Ensure GIDClientID is set in Info.plist

### Apple Sign-In not working

- Verify Sign in with Apple capability is added
- Check entitlements file is configured
- Ensure provisioning profile includes capability

### Network errors

- For simulator, localhost should work
- For physical device, use your machine's IP address
- Ensure backend is running on correct port

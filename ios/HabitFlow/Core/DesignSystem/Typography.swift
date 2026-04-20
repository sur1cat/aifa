import SwiftUI

extension Font {
    static let hf = HFTypography()
}

struct HFTypography {
    let largeTitle = Font.system(size: 34, weight: .bold, design: .rounded)
    let title1 = Font.system(size: 28, weight: .bold, design: .rounded)
    let title2 = Font.system(size: 22, weight: .bold, design: .rounded)
    let title3 = Font.system(size: 20, weight: .semibold, design: .rounded)
    let headline = Font.system(size: 17, weight: .semibold, design: .rounded)
    let body = Font.system(size: 17, weight: .regular, design: .rounded)
    let callout = Font.system(size: 16, weight: .regular, design: .rounded)
    let subheadline = Font.system(size: 15, weight: .regular, design: .rounded)
    let footnote = Font.system(size: 13, weight: .regular, design: .rounded)
    let caption = Font.system(size: 12, weight: .regular, design: .rounded)
    let caption2 = Font.system(size: 11, weight: .regular, design: .rounded)
}

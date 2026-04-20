import SwiftUI

extension Font {
    static let hf = HFTypography()
}

struct HFTypography {
    // MARK: - Display
    let display = Font.system(size: 42, weight: .bold, design: .rounded)
    let largeTitle = Font.system(size: 34, weight: .bold, design: .rounded)
    let title1 = Font.system(size: 28, weight: .bold, design: .rounded)
    let title2 = Font.system(size: 22, weight: .bold, design: .rounded)
    let title3 = Font.system(size: 20, weight: .semibold, design: .rounded)

    // MARK: - Body
    let headline = Font.system(size: 17, weight: .semibold, design: .rounded)
    let body = Font.system(size: 17, weight: .regular, design: .rounded)
    let bodyBold = Font.system(size: 17, weight: .bold, design: .rounded)
    let callout = Font.system(size: 16, weight: .regular, design: .rounded)
    let subheadline = Font.system(size: 15, weight: .regular, design: .rounded)
    let subheadlineBold = Font.system(size: 15, weight: .semibold, design: .rounded)

    // MARK: - Small
    let footnote = Font.system(size: 13, weight: .regular, design: .rounded)
    let footnoteBold = Font.system(size: 13, weight: .semibold, design: .rounded)
    let caption = Font.system(size: 12, weight: .regular, design: .rounded)
    let captionBold = Font.system(size: 12, weight: .semibold, design: .rounded)
    let micro = Font.system(size: 11, weight: .regular, design: .rounded)

    // MARK: - Numbers
    let numberLarge = Font.system(size: 36, weight: .bold, design: .rounded)
    let numberMedium = Font.system(size: 28, weight: .semibold, design: .rounded)
    let numberSmall = Font.system(size: 20, weight: .semibold, design: .rounded)
}

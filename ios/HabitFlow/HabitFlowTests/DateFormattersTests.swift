import XCTest
@testable import HabitFlow

final class DateFormattersTests: XCTestCase {

    func testApiDateFormat() {
        let date = DateComponents(
            calendar: Calendar.current,
            year: 2024,
            month: 1,
            day: 15
        ).date!

        let formatted = DateFormatters.apiDate.string(from: date)
        XCTAssertEqual(formatted, "2024-01-15")
    }

    func testApiDateParse() {
        let dateString = "2024-06-20"
        let date = DateFormatters.apiDate.date(from: dateString)

        XCTAssertNotNil(date)

        let calendar = Calendar.current
        XCTAssertEqual(calendar.component(.year, from: date!), 2024)
        XCTAssertEqual(calendar.component(.month, from: date!), 6)
        XCTAssertEqual(calendar.component(.day, from: date!), 20)
    }

    func testTime24hFormat() {
        var components = DateComponents()
        components.hour = 14
        components.minute = 30
        let date = Calendar.current.date(from: components)!

        let formatted = DateFormatters.time24h.string(from: date)
        XCTAssertEqual(formatted, "14:30")
    }

    func testShortMonthDayFormat() {
        let date = DateComponents(
            calendar: Calendar.current,
            year: 2024,
            month: 3,
            day: 5
        ).date!

        let formatted = DateFormatters.shortMonthDay.string(from: date)
        XCTAssertEqual(formatted, "Mar 5")
    }

    func testMonthYearFormat() {
        let date = DateComponents(
            calendar: Calendar.current,
            year: 2024,
            month: 12,
            day: 1
        ).date!

        let formatted = DateFormatters.monthYear.string(from: date)
        XCTAssertEqual(formatted, "December 2024")
    }

    func testShortWeekdayFormat() {
        // Create a known date (Monday)
        let date = DateComponents(
            calendar: Calendar.current,
            year: 2024,
            month: 1,
            day: 15 // This is a Monday
        ).date!

        let formatted = DateFormatters.shortWeekday.string(from: date)
        XCTAssertEqual(formatted, "Mon")
    }

    func testDayOfMonthFormat() {
        let date = DateComponents(
            calendar: Calendar.current,
            year: 2024,
            month: 1,
            day: 7
        ).date!

        let formatted = DateFormatters.dayOfMonth.string(from: date)
        XCTAssertEqual(formatted, "7")
    }

    func testFormatterReuse() {
        // Verify that the same formatter instance is returned
        let formatter1 = DateFormatters.apiDate
        let formatter2 = DateFormatters.apiDate

        XCTAssertTrue(formatter1 === formatter2, "Formatters should be the same instance")
    }

    func testISO8601Format() {
        let dateString = "2024-01-15T10:30:00Z"
        let date = DateFormatters.iso8601Basic.date(from: dateString)

        XCTAssertNotNil(date)
    }

    func testAllFormattersAreNotNil() {
        XCTAssertNotNil(DateFormatters.apiDate)
        XCTAssertNotNil(DateFormatters.iso8601)
        XCTAssertNotNil(DateFormatters.iso8601Basic)
        XCTAssertNotNil(DateFormatters.time24h)
        XCTAssertNotNil(DateFormatters.time12h)
        XCTAssertNotNil(DateFormatters.shortMonthDay)
        XCTAssertNotNil(DateFormatters.monthYear)
        XCTAssertNotNil(DateFormatters.fullDate)
        XCTAssertNotNil(DateFormatters.mediumDate)
        XCTAssertNotNil(DateFormatters.weekday)
        XCTAssertNotNil(DateFormatters.shortWeekday)
        XCTAssertNotNil(DateFormatters.dayOfMonth)
        XCTAssertNotNil(DateFormatters.shortMonth)
        XCTAssertNotNil(DateFormatters.fullMonth)
        XCTAssertNotNil(DateFormatters.year)
        XCTAssertNotNil(DateFormatters.dateTime)
    }
}

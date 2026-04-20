import XCTest
@testable import HabitFlow

final class LRUCacheTests: XCTestCase {

    func testBasicSetAndGet() {
        let cache = LRUCache<String, Int>(capacity: 3)

        cache.set("a", value: 1)
        cache.set("b", value: 2)
        cache.set("c", value: 3)

        XCTAssertEqual(cache.get("a"), 1)
        XCTAssertEqual(cache.get("b"), 2)
        XCTAssertEqual(cache.get("c"), 3)
    }

    func testGetReturnsNilForMissingKey() {
        let cache = LRUCache<String, Int>(capacity: 3)

        XCTAssertNil(cache.get("missing"))
    }

    func testEvictsLRUWhenAtCapacity() {
        let cache = LRUCache<String, Int>(capacity: 2)

        cache.set("a", value: 1)
        cache.set("b", value: 2)
        // Cache is now full

        cache.set("c", value: 3)
        // "a" should be evicted (LRU)

        XCTAssertNil(cache.get("a"), "LRU item 'a' should be evicted")
        XCTAssertEqual(cache.get("b"), 2)
        XCTAssertEqual(cache.get("c"), 3)
    }

    func testAccessUpdatesLRUOrder() {
        let cache = LRUCache<String, Int>(capacity: 2)

        cache.set("a", value: 1)
        cache.set("b", value: 2)

        // Access "a" to make it recently used
        _ = cache.get("a")

        // Add new item - "b" should be evicted (now LRU)
        cache.set("c", value: 3)

        XCTAssertEqual(cache.get("a"), 1, "'a' was accessed, should not be evicted")
        XCTAssertNil(cache.get("b"), "'b' was LRU, should be evicted")
        XCTAssertEqual(cache.get("c"), 3)
    }

    func testUpdateExistingKey() {
        let cache = LRUCache<String, Int>(capacity: 2)

        cache.set("a", value: 1)
        cache.set("a", value: 10)

        XCTAssertEqual(cache.get("a"), 10, "Value should be updated")
        XCTAssertEqual(cache.count, 1, "Count should not increase for update")
    }

    func testRemove() {
        let cache = LRUCache<String, Int>(capacity: 3)

        cache.set("a", value: 1)
        cache.set("b", value: 2)

        cache.remove("a")

        XCTAssertNil(cache.get("a"))
        XCTAssertEqual(cache.get("b"), 2)
        XCTAssertEqual(cache.count, 1)
    }

    func testClear() {
        let cache = LRUCache<String, Int>(capacity: 3)

        cache.set("a", value: 1)
        cache.set("b", value: 2)
        cache.set("c", value: 3)

        cache.clear()

        XCTAssertNil(cache.get("a"))
        XCTAssertNil(cache.get("b"))
        XCTAssertNil(cache.get("c"))
        XCTAssertEqual(cache.count, 0)
    }

    func testCount() {
        let cache = LRUCache<String, Int>(capacity: 5)

        XCTAssertEqual(cache.count, 0)

        cache.set("a", value: 1)
        XCTAssertEqual(cache.count, 1)

        cache.set("b", value: 2)
        XCTAssertEqual(cache.count, 2)

        cache.remove("a")
        XCTAssertEqual(cache.count, 1)
    }

    func testCapacityOfOne() {
        let cache = LRUCache<String, Int>(capacity: 1)

        cache.set("a", value: 1)
        XCTAssertEqual(cache.get("a"), 1)

        cache.set("b", value: 2)
        XCTAssertNil(cache.get("a"), "'a' should be evicted")
        XCTAssertEqual(cache.get("b"), 2)
    }

    func testThreadSafety() {
        let cache = LRUCache<Int, Int>(capacity: 100)
        let expectation = XCTestExpectation(description: "Concurrent operations")
        expectation.expectedFulfillmentCount = 10

        for i in 0..<10 {
            DispatchQueue.global().async {
                for j in 0..<100 {
                    cache.set(i * 100 + j, value: j)
                    _ = cache.get(i * 100 + j)
                }
                expectation.fulfill()
            }
        }

        wait(for: [expectation], timeout: 5.0)
        // If we get here without crash, thread safety works
    }
}

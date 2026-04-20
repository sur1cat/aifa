import Foundation
import os

/// Thread-safe LRU (Least Recently Used) cache implementation
/// Automatically evicts least recently used items when capacity is exceeded
final class LRUCache<Key: Hashable, Value>: @unchecked Sendable {
    private let capacity: Int
    private var cache: [Key: Value] = [:]
    private var order: [Key] = []
    private var lockPtr: UnsafeMutablePointer<os_unfair_lock>

    init(capacity: Int) {
        self.capacity = max(1, capacity)
        self.lockPtr = UnsafeMutablePointer<os_unfair_lock>.allocate(capacity: 1)
        self.lockPtr.initialize(to: os_unfair_lock())
    }

    deinit {
        lockPtr.deinitialize(count: 1)
        lockPtr.deallocate()
    }

    private func withLock<T>(_ block: () -> T) -> T {
        os_unfair_lock_lock(lockPtr)
        defer { os_unfair_lock_unlock(lockPtr) }
        return block()
    }

    /// Get value for key, returns nil if not found
    func get(_ key: Key) -> Value? {
        withLock {
            guard let value = cache[key] else { return nil }
            // Move to end (most recently used)
            if let index = order.firstIndex(of: key) {
                order.remove(at: index)
                order.append(key)
            }
            return value
        }
    }

    /// Set value for key, evicting LRU item if at capacity
    func set(_ key: Key, value: Value) {
        withLock {
            if cache[key] != nil {
                // Update existing - move to end
                if let index = order.firstIndex(of: key) {
                    order.remove(at: index)
                }
            } else if cache.count >= capacity {
                // Evict LRU (first item)
                if let lruKey = order.first {
                    cache.removeValue(forKey: lruKey)
                    order.removeFirst()
                }
            }

            cache[key] = value
            order.append(key)
        }
    }

    /// Remove value for key
    func remove(_ key: Key) {
        withLock {
            cache.removeValue(forKey: key)
            if let index = order.firstIndex(of: key) {
                order.remove(at: index)
            }
        }
    }

    /// Clear all cached values
    func clear() {
        withLock {
            cache.removeAll()
            order.removeAll()
        }
    }

    /// Current number of cached items
    var count: Int {
        withLock {
            return cache.count
        }
    }
}

// MARK: - Application Caches

/// Shared caches for common data
enum AppCache {
    /// Cache for formatted currency strings (key: "\(amount)_\(currency)")
    static let currencyStrings = LRUCache<String, String>(capacity: 100)

    /// Cache for formatted date strings (key: "\(date.timeIntervalSince1970)_\(format)")
    static let dateStrings = LRUCache<String, String>(capacity: 200)

    /// Cache for habit completion rates (key: habitID)
    static let habitCompletionRates = LRUCache<UUID, Double>(capacity: 50)

    /// Cache for weekly stats (key: "week_\(weekNumber)_\(year)")
    static let weeklyStats = LRUCache<String, WeeklyStats>(capacity: 10)

    /// Clear all caches (call on logout or data refresh)
    static func clearAll() {
        currencyStrings.clear()
        dateStrings.clear()
        habitCompletionRates.clear()
        weeklyStats.clear()
    }
}

// MARK: - Weekly Stats

struct WeeklyStats: Sendable {
    let habitsCompletionRate: Double
    let tasksCompleted: Int
    let tasksTotal: Int
    let totalExpenses: Double
    let totalIncome: Double
}

// MARK: - Convenience Extensions

extension DataManager {
    /// Cached currency formatting
    func formatCurrencyCached(_ amount: Double) -> String {
        let key = "\(amount)_\(profile.currency.rawValue)"

        if let cached = AppCache.currencyStrings.get(key) {
            return cached
        }

        let formatted = formatCurrency(amount)
        AppCache.currencyStrings.set(key, value: formatted)
        return formatted
    }
}

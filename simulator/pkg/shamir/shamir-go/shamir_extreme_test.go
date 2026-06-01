package shamir

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func monitorGC() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastGC time.Time
	for {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			if !lastGC.IsZero() {
				gcInterval := time.Since(lastGC)
				fmt.Printf("GC interval: %.2f seconds, Allocated: %.2f MB, GC cycles: %d\n",
					gcInterval.Seconds(),
					float64(memStats.Alloc)/1024/1024,
					memStats.NumGC)
				if gcInterval < 1*time.Second {
					fmt.Println("WARNING: GC interval is less than 1 second (potential卡顿)")
				}
			}
			lastGC = time.Now()
		}
	}
}

func testFieldSize(prime int64, numSecrets int) bool {
	fmt.Printf("\nTesting field size: %d, secrets: %d\n", prime, numSecrets)

	start := time.Now()
	scheme, err := NewShamir(3, 5, prime)
	if err != nil {
		fmt.Printf("Error creating scheme: %v\n", err)
		return false
	}

	var sharesList [][][2]int64
	for i := 0; i < numSecrets; i++ {
		secret := int64(i) % (prime - 1)
		shares, err := scheme.Split(secret)
		if err != nil {
			fmt.Printf("Error splitting secret: %v\n", err)
			return false
		}
		sharesList = append(sharesList, shares)
	}

	for i, shares := range sharesList {
		recovered, err := scheme.Recover(shares[:3])
		if err != nil {
			fmt.Printf("Error recovering secret %d: %v\n", i, err)
			return false
		}
		expected := int64(i) % (scheme.GetPrime() - 1)
		if recovered != expected {
			fmt.Printf("Recovery failed: expected %d, got %d\n", expected, recovered)
			return false
		}
	}

	duration := time.Since(start)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Printf("Success! Time: %.2f ms, Allocated: %.2f MB, GC cycles: %d\n",
		float64(duration.Milliseconds()),
		float64(memStats.Alloc)/1024/1024,
		memStats.NumGC)

	return true
}

func TestExtremeFieldSizes(t *testing.T) {
	// 测试不同域大小（使用真正的素数）
	fieldSizes := []int64{257, 1031, 4099, 16381, 65537, 262147, 1048573}
	secretCounts := []int{100, 500, 1000, 5000, 10000, 50000}

	fmt.Println("Starting extreme performance test...")
	fmt.Println("Monitoring GC every 1 second...")

	go monitorGC()

	for _, fieldSize := range fieldSizes {
		for _, secretCount := range secretCounts {
			success := testFieldSize(fieldSize, secretCount)
			if !success {
				fmt.Printf("Test failed for field size %d, secrets %d\n", fieldSize, secretCount)
				break
			}
			time.Sleep(2 * time.Second) // 让GC有时间运行
		}
	}

	fmt.Println("\nTest completed!")
}

package main

import "testing"

func Test_verifyWorkerCount(t *testing.T) {
	tests := []struct {
		loadTotal int
		workers   int
		expected  int
	}{
		{1, 2, 1},
		{3, 4, 3},
		{4, 3, 3},
		{10000, 10, 10},
	}

	for _, test := range tests {
		workers := verifyWorkerCount(test.loadTotal, test.workers)
		if workers != test.expected {
			t.Fatalf("verifyWorkerCount failed. Expected: %d, received: %d", test.expected, workers)
		}
	}
}

func Test_calculateLoad(t *testing.T) {
	tests := []struct {
		totalLoad int
		workers   int
		turn      int
		start     int
		end       int
	}{
		{3, 3, 0, 0, 1},
		{3, 3, 1, 1, 2},
		{3, 3, 2, 2, 3},
		{100, 3, 0, 0, 33},
		{100, 3, 1, 33, 66},
		{100, 3, 2, 66, 100},
		{37, 2, 0, 0, 18},
		{37, 2, 1, 18, 19},
	}

	for _, test := range tests {
		start, end := calculateLoad(test.totalLoad, test.workers, test.turn)
		if start != test.start && end != test.end {
			t.Fatal("loads did not calculate for:", test)
		}
	}
}

func Test_available(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"jesselucas", false},
		{"forestgiant", false},
		{"madeupusernamenoonewillhave", true},
	}

	auth, cookie := getAuth()
	// fmt.Println(cookie)
	for _, test := range tests {
		ok := available(test.name, auth, cookie)
		if ok != test.expected {
			t.Fatalf("expected: %t, received: %t, for name: %s", test.expected, ok, test.name)
		}
	}
}

func Test_splitData(t *testing.T) {
	tests := []struct {
		data     []byte
		expected []string
	}{
		{[]byte("jesselucas\n forestgiant"), []string{"jesselucas", "forestgiant"}},
		{[]byte("jesselucas\n forestgiant\n"), []string{"jesselucas", "forestgiant"}},
		{[]byte("jesselucas\n \n \n forestgiant\n \n"), []string{"jesselucas", "forestgiant"}},
	}

	for _, test := range tests {
		names, err := splitData(test.data)
		if err != nil || len(names) != len(test.expected) {
			t.Fatal("splitData failed")
		}
	}
}

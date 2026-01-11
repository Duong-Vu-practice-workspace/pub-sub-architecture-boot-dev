package __publisher

import (
	"fmt"
	"slices"
	"testing"
)

func Test(t *testing.T) {
	type testCase struct {
		users                  []user
		mv                     move
		expectedFightLocations []piece
	}

	runCases := []testCase{
		{
			users: []user{
				{
					name: "Toussaint",
					pieces: []piece{
						{
							location: "San Domingo",
							name:     "Cavalry",
						},
						{
							location: "San Domingo",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Napoleon",
					pieces: []piece{
						{
							location: "France",
							name:     "Infantry",
						},
						{
							location: "Russia",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Washington",
					pieces: []piece{
						{
							location: "United States",
							name:     "Artillery",
						},
					},
				},
			},
			mv: move{
				userName: "Toussaint",
				piece: piece{
					location: "United States",
					name:     "Cavalry",
				},
			},
			expectedFightLocations: []piece{
				{
					location: "United States",
					name:     "Artillery",
				},
			},
		},
	}

	submitCases := append(runCases, []testCase{
		{
			users: []user{
				{
					name: "Toussaint",
					pieces: []piece{
						{
							location: "San Domingo",
							name:     "Cavalry",
						},
						{
							location: "San Domingo",
							name:     "Infantry",
						},
					},
				},
				{
					name: "Napoleon",
					pieces: []piece{
						{
							location: "France",
							name:     "Infantry",
						},
						{
							location: "Russia",
							name:     "Infantry",
						},
						{
							location: "United States",
							name:     "Cavalry",
						},
					},
				},
				{
					name: "Washington",
					pieces: []piece{
						{
							location: "United States",
							name:     "Artillery",
						},
					},
				},
			},
			mv: move{
				userName: "Toussaint",
				piece: piece{
					location: "United States",
					name:     "Cavalry",
				},
			},
			expectedFightLocations: []piece{
				{
					location: "United States",
					name:     "Cavalry",
				},
				{
					location: "United States",
					name:     "Artillery",
				},
			},
		},
	}...)

	testCases := runCases
	if withSubmit {
		testCases = submitCases
	}

	skipped := len(submitCases) - len(testCases)

	passed, failed := 0, 0

	for _, test := range testCases {
		bufferedCh := make(chan move, 100)
		mover := user{}

		for _, u := range test.users {
			if u.name == test.mv.userName {
				mover = u
			}
		}
		if mover.name == "" {
			t.Errorf("Test Failed: user with name %v not found", test.mv.userName)
			failed++
			continue
		}
		mover.march(test.mv.piece, bufferedCh)
		close(bufferedCh)
		output := doBattles(bufferedCh, test.users)
		if !slices.Equal(output, test.expectedFightLocations) {
			t.Errorf(`---------------------------------
Test Failed:
  users:
%v
  move: %v
  =>
  expected battle pieces:
%v
  actual battle pieces:
%v
`,
				formatSlice(test.users),
				test.mv,
				formatSlice(test.expectedFightLocations),
				formatSlice(output),
			)
			failed++
		} else {
			fmt.Printf(`---------------------------------
Test Passed:
  users:
%v
  move: %v
  =>
  expected battle pieces:
%v
  actual battle pieces:
%v
`,
				formatSlice(test.users),
				test.mv,
				formatSlice(test.expectedFightLocations),
				formatSlice(output),
			)
			passed++
		}
	}

	fmt.Println("---------------------------------")
	if skipped > 0 {
		fmt.Printf("%d passed, %d failed, %d skipped\n", passed, failed, skipped)
	} else {
		fmt.Printf("%d passed, %d failed\n", passed, failed)
	}
}

func formatSlice[T any](slice []T) string {
	if slice == nil {
		return "nil"
	}
	output := ""
	for i, v := range slice {
		output += fmt.Sprintf("* %v", v)
		if i < len(slice)-1 {
			output += "\n"
		}
	}
	return output
}

// withSubmit is set at compile time depending
// on which button is used to run the tests
var withSubmit = true

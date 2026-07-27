package feeds

import (
	"testing"
	"time"
)

const realDeletionMessage = "CAR 55 (SAI) TIME 1:20.046 DELETED - TRACK LIMITS AT TURN 2 LAP 3 15:06:13"

func lapTimeOf(s string) time.Duration {
	d, err := parseF1Duration(s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestMarkDeletedLaps_RealMessageText_MarksMatchingLap(t *testing.T) {
	deletedTime := lapTimeOf("1:20.046")
	laps := map[string][]LapData{
		"55": {
			{LapNo: 3, LapTime: &deletedTime, IsPersonalBest: true},
			{LapNo: 4, LapTime: durPtr(1*time.Minute + 21*time.Second)},
		},
	}
	messages := []RaceControlMessage{{Message: realDeletionMessage}}

	MarkDeletedLaps(laps, messages)

	got := laps["55"][0]
	if !got.Deleted {
		t.Fatalf("lap 3: Deleted = false, want true")
	}
	if got.IsPersonalBest {
		t.Errorf("lap 3: IsPersonalBest = true, want false (forced off on deletion)")
	}
	if want := "TRACK LIMITS AT TURN 2 LAP 3"; got.DeletedReason != want {
		t.Errorf("lap 3: DeletedReason = %q, want %q (local track timestamp stripped)", got.DeletedReason, want)
	}
	if laps["55"][1].Deleted {
		t.Errorf("lap 4: Deleted = true, want false (different LapTime, shouldn't match)")
	}
}

func TestMarkDeletedLaps_ReinstatedLap_NeverMarkedDeleted(t *testing.T) {
	deletedTime := lapTimeOf("1:20.046")
	laps := map[string][]LapData{
		"55": {{LapNo: 3, LapTime: &deletedTime, IsPersonalBest: true}},
	}
	messages := []RaceControlMessage{
		{Message: realDeletionMessage},
		{Message: "CAR 55 (SAI) TIME 1:20.046 LAP DELETION TIME REINSTATED"},
	}

	MarkDeletedLaps(laps, messages)

	got := laps["55"][0]
	if got.Deleted {
		t.Errorf("Deleted = true, want false (reinstated later)")
	}
	if got.DeletedReason != "" {
		t.Errorf("DeletedReason = %q, want empty (reinstated later)", got.DeletedReason)
	}
	if !got.IsPersonalBest {
		t.Errorf("IsPersonalBest = false, want true (untouched — never actually marked deleted)")
	}
}

func TestMarkDeletedLaps_ReinstatementBeforeDeletionInStream_StillSuppressed(t *testing.T) {

	deletedTime := lapTimeOf("1:20.046")
	laps := map[string][]LapData{
		"55": {{LapNo: 3, LapTime: &deletedTime}},
	}
	messages := []RaceControlMessage{
		{Message: "CAR 55 (SAI) TIME 1:20.046 LAP DELETION TIME REINSTATED"},
		{Message: realDeletionMessage},
	}

	MarkDeletedLaps(laps, messages)

	if laps["55"][0].Deleted {
		t.Errorf("Deleted = true, want false (reinstated)")
	}
}

func TestMarkDeletedLaps_NoMessagesIsNoop(t *testing.T) {
	laps := map[string][]LapData{"1": {{Timestamp: time.Minute}}}
	MarkDeletedLaps(laps, nil)
	if laps["1"][0].Deleted {
		t.Errorf("Deleted = true, want false (no messages at all)")
	}
}

func TestMarkDeletedLaps_NonMatchingDriverOrTime_Untouched(t *testing.T) {
	otherTime := 1*time.Minute + 22*time.Second
	laps := map[string][]LapData{
		"44": {{LapNo: 1, LapTime: &otherTime}},
	}
	messages := []RaceControlMessage{{Message: realDeletionMessage}}

	MarkDeletedLaps(laps, messages)

	if laps["44"][0].Deleted {
		t.Errorf("Deleted = true, want false (different driver number)")
	}
}

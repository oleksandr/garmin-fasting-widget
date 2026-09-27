import Toybox.Application.Storage;
import Toybox.Application.Properties;
import Toybox.Time;
import Toybox.System;
import Toybox.Lang;

(:glance)
module FastingStorage {

    enum State {
        STATE_IDLE = 0,
        STATE_FASTING = 1,
        STATE_GOAL_REACHED = 2,
        STATE_EATING = 3
    }

    const KEY_STATE = "fasting_state";
    const KEY_START_EPOCH = "fasting_start_epoch";
    const KEY_TARGET_HOURS = "fasting_target_hours";
    const KEY_HISTORY = "fasting_history";
    const MAX_HISTORY_ITEMS = 30;

    function getState() as Number {
        var val = Storage.getValue(KEY_STATE);
        return (val != null) ? (val as Number) : STATE_IDLE;
    }

    function setState(newState as Number) as Void {
        Storage.setValue(KEY_STATE, newState);
    }

    function getStartEpoch() as Number {
        var val = Storage.getValue(KEY_START_EPOCH);
        return (val != null) ? (val as Number) : 0;
    }

    function setStartEpoch(epoch as Number) as Void {
        Storage.setValue(KEY_START_EPOCH, epoch);
    }

    function shiftStartEpoch(secondsDelta as Number) as Void {
        var current = getStartEpoch();
        if (current > 0) {
            setStartEpoch(current + secondsDelta);
        }
    }

    function getTargetHours() as Number {
        try {
            var val = Properties.getValue("targetFastHours");
            if (val != null) {
                return val as Number;
            }
        } catch (e) {
            // fallback
        }
        return 16;
    }

    function getEatingHours() as Number {
        try {
            var val = Properties.getValue("targetEatHours");
            if (val != null) {
                return val as Number;
            }
        } catch (e) {
            // fallback
        }
        return 8;
    }

    function startFast() as Void {
        var now = Time.now().value();
        setStartEpoch(now);
        setState(STATE_FASTING);
    }

    function endFastAndStartEating() as Void {
        var start = getStartEpoch();
        var now = Time.now().value();
        var target = getTargetHours();
        var durationHours = (now - start) / 3600.0;
        var stateStr = (durationHours >= target) ? "COMPLETED" : "SHORT";

        addHistoryRecord(start, now, target, durationHours, stateStr);

        // Switch to eating window
        setStartEpoch(now);
        setState(STATE_EATING);
    }

    function endEatingAndStartFast() as Void {
        startFast();
    }

    function reset() as Void {
        setState(STATE_IDLE);
        setStartEpoch(0);
    }

    function getHistory() as Array<Dictionary> {
        var hist = Storage.getValue(KEY_HISTORY);
        if (hist != null && hist instanceof Array) {
            return hist as Array<Dictionary>;
        }
        return [];
    }

    function addHistoryRecord(start as Number, endVal as Number, target as Number, duration as Float, stateStr as String) as Void {
        var history = getHistory();
        var record = {
            "id" => start,
            "start" => start,
            "end" => endVal,
            "target_hours" => target,
            "duration_hours" => duration,
            "state" => stateStr,
            "synced" => false
        };

        history.add(record);

        // Keep last 30 items
        if (history.size() > MAX_HISTORY_ITEMS) {
            history = history.slice(history.size() - MAX_HISTORY_ITEMS, null);
        }

        Storage.setValue(KEY_HISTORY, history);
    }

    function getUnsyncedFasts() as Array<Dictionary> {
        var history = getHistory();
        var unsynced = [];
        for (var i = 0; i < history.size(); i++) {
            var item = history[i];
            var isSynced = item.get("synced") as Boolean?;
            if (isSynced == null || isSynced == false) {
                unsynced.add(item);
            }
        }
        return unsynced;
    }

    function markFastsSynced(syncedIds as Array<Number>) as Void {
        var history = getHistory();
        for (var i = 0; i < history.size(); i++) {
            var item = history[i];
            var id = item.get("id") as Number?;
            if (id != null && syncedIds.indexOf(id) != -1) {
                item.put("synced", true);
            }
        }
        Storage.setValue(KEY_HISTORY, history);
    }
}

import Toybox.WatchUi;
import Toybox.System;
import Toybox.Attention;
import Toybox.Lang;

class FastingMenuDelegate extends WatchUi.BehaviorDelegate {

    private var _view as FastingMainView;

    function initialize(view as FastingMainView) {
        BehaviorDelegate.initialize();
        _view = view;
    }

    function onKey(keyEvent as WatchUi.KeyEvent) as Boolean {
        var key = keyEvent.getKey();

        if (key == WatchUi.KEY_UP) {
            // Shift start time earlier by 15 mins (adds 15m to elapsed)
            FastingStorage.shiftStartEpoch(-900);
            vibrateShort();
            _view.showStatus("+15 min added");
            return true;
        } else if (key == WatchUi.KEY_DOWN) {
            // Shift start time later by 15 mins (subtracts 15m from elapsed)
            FastingStorage.shiftStartEpoch(900);
            vibrateShort();
            _view.showStatus("-15 min adjusted");
            return true;
        }

        return false;
    }

    function onSelect() as Boolean {
        vibrateShort();
        openActionMenu();
        return true;
    }

    private function openActionMenu() as Void {
        var state = FastingStorage.getState();
        var menu = new WatchUi.Menu2({:title => "Fasting Actions"});

        if (state == FastingStorage.STATE_IDLE) {
            menu.addItem(new WatchUi.MenuItem("Start Fast", "Begin 16/8 fasting", :start_fast, null));
            menu.addItem(new WatchUi.MenuItem("Sync Now", "Push history to API", :sync_api, null));
        } else if (state == FastingStorage.STATE_FASTING) {
            menu.addItem(new WatchUi.MenuItem("End Fast & Eat", "Start eating window", :end_fast, null));
            menu.addItem(new WatchUi.MenuItem("+1 Hour Fasted", "Started 1h earlier", :shift_plus_60, null));
            menu.addItem(new WatchUi.MenuItem("-1 Hour Fasted", "Started 1h later", :shift_minus_60, null));
            menu.addItem(new WatchUi.MenuItem("Sync Now", "Push history to API", :sync_api, null));
            menu.addItem(new WatchUi.MenuItem("Reset / Cancel", "Discard active fast", :reset_fast, null));
        } else if (state == FastingStorage.STATE_EATING) {
            menu.addItem(new WatchUi.MenuItem("Start Fasting", "Close eating window", :start_fast, null));
            menu.addItem(new WatchUi.MenuItem("Sync Now", "Push history to API", :sync_api, null));
            menu.addItem(new WatchUi.MenuItem("Reset / Cancel", "Discard eating window", :reset_fast, null));
        }

        var delegate = new FastingMenuActionDelegate(_view);
        WatchUi.pushView(menu, delegate, WatchUi.SLIDE_UP);
    }

    private function vibrateShort() as Void {
        if (Attention has :vibrate) {
            try {
                Attention.vibrate([new Attention.VibeProfile(50, 100)]);
            } catch (e) {
                // Ignore if not supported on simulator
            }
        }
    }
}

class FastingMenuActionDelegate extends WatchUi.Menu2InputDelegate {

    private var _view as FastingMainView;

    function initialize(view as FastingMainView) {
        Menu2InputDelegate.initialize();
        _view = view;
    }

    function onSelect(item as WatchUi.MenuItem) as Void {
        var id = item.getId();
        WatchUi.popView(WatchUi.SLIDE_DOWN);

        if (id == :start_fast) {
            FastingStorage.startFast();
            _view.showStatus("Fast Started!");
        } else if (id == :end_fast) {
            FastingStorage.endFastAndStartEating();
            _view.showStatus("Eating Window Started");
        } else if (id == :shift_plus_60) {
            FastingStorage.shiftStartEpoch(-3600);
            _view.showStatus("+1 Hour Added");
        } else if (id == :shift_minus_60) {
            FastingStorage.shiftStartEpoch(3600);
            _view.showStatus("-1 Hour Adjusted");
        } else if (id == :reset_fast) {
            FastingStorage.reset();
            _view.showStatus("Fast Reset");
        } else if (id == :sync_api) {
            _view.showStatus("Syncing to LAN...");
            var sync = new FastingSync(method(:onSyncResult));
            sync.startSync();
        }
    }

    function onSyncResult(success as Boolean, message as String) as Void {
        _view.showStatus(message);
    }

    function onBack() as Void {
        WatchUi.popView(WatchUi.SLIDE_DOWN);
    }
}

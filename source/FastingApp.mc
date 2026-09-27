import Toybox.Application;
import Toybox.WatchUi;
import Toybox.Background;
import Toybox.Time;
import Toybox.System;
import Toybox.Lang;

class FastingApp extends Application.AppBase {

    function initialize() {
        AppBase.initialize();
    }

    function onStart(state as Dictionary?) as Void {
    }

    function onStop(state as Dictionary?) as Void {
        // Schedule background reminder when leaving the widget
        try {
            var fastState = FastingStorage.getState();
            if (fastState == FastingStorage.STATE_FASTING) {
                var start = FastingStorage.getStartEpoch();
                var target = FastingStorage.getTargetHours();
                var targetEpoch = start + (target * 3600);
                var now = Time.now().value();

                if (targetEpoch > now) {
                    var wakeMoment = new Time.Moment(targetEpoch);
                    Background.registerForTemporalEvent(wakeMoment);
                }
            } else if (fastState == FastingStorage.STATE_EATING) {
                var start = FastingStorage.getStartEpoch();
                var eatHours = FastingStorage.getEatingHours();
                var targetEpoch = start + (eatHours * 3600);
                var now = Time.now().value();

                if (targetEpoch > now) {
                    var wakeMoment = new Time.Moment(targetEpoch);
                    Background.registerForTemporalEvent(wakeMoment);
                }
            }
        } catch (e) {
            System.println("Background registration error: " + e.getErrorMessage());
        }
    }

    function getInitialView() as [WatchUi.Views] or [WatchUi.Views, WatchUi.InputDelegates] {
        var view = new FastingMainView();
        var delegate = new FastingMenuDelegate(view);
        return [view, delegate];
    }

    function getGlanceView() as [WatchUi.GlanceView] or [WatchUi.GlanceView, WatchUi.GlanceViewDelegate] or Null {
        return [new FastingGlanceView()];
    }

    function getServiceDelegate() as [System.ServiceDelegate] {
        return [new FastingServiceDelegate()];
    }
}

function getApp() as FastingApp {
    return Application.getApp() as FastingApp;
}

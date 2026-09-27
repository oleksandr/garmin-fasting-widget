import Toybox.WatchUi;
import Toybox.Graphics;
import Toybox.Time;
import Toybox.Time.Gregorian;
import Toybox.Timer;
import Toybox.System;
import Toybox.Lang;

class FastingMainView extends WatchUi.View {

    private var _timer as Timer.Timer?;
    private var _statusBanner as String?;
    private var _bannerTicks as Number = 0;

    function initialize() {
        View.initialize();
    }

    function onShow() as Void {
        _timer = new Timer.Timer();
        _timer.start(method(:onTimerTick), 1000, true);
    }

    function onHide() as Void {
        if (_timer != null) {
            _timer.stop();
            _timer = null;
        }
    }

    function onTimerTick() as Void {
        if (_bannerTicks > 0) {
            _bannerTicks--;
            if (_bannerTicks == 0) {
                _statusBanner = null;
            }
        }
        WatchUi.requestUpdate();
    }

    function showStatus(msg as String) as Void {
        _statusBanner = msg;
        _bannerTicks = 3; // Show for 3 seconds
        WatchUi.requestUpdate();
    }

    function onUpdate(dc as Graphics.Dc) as Void {
        var w = dc.getWidth();
        var h = dc.getHeight();
        var cx = w / 2;
        var cy = h / 2;
        var radius = (w / 2) - 14;

        // Pure black AMOLED background
        dc.setColor(Graphics.COLOR_BLACK, Graphics.COLOR_BLACK);
        dc.clear();

        var state = FastingStorage.getState();
        var start = FastingStorage.getStartEpoch();
        var now = Time.now().value();

        var stateTitle = "FASTING " + FastingStorage.getTargetHours() + "/" + FastingStorage.getEatingHours();
        var elapsedInfo = "Ready to start fast";
        var timerDigits = "00:00:00";
        var timerSub = "PRESS GPS TO START";
        var targetInfo = "Goal: " + FastingStorage.getTargetHours() + "h Fasting";
        var arcColor = 0x00D2FF; // Cyan
        var progress = 0.0;

        if (state == FastingStorage.STATE_FASTING) {
            var targetHours = FastingStorage.getTargetHours();
            var targetSec = targetHours * 3600;
            var elapsedSec = (start > 0 && now >= start) ? (now - start) : 0;
            var elapsedH = elapsedSec / 3600;
            var elapsedM = (elapsedSec % 3600) / 60;
            var targetEpoch = start + targetSec;

            if (elapsedSec >= targetSec) {
                stateTitle = "GOAL REACHED!";
                elapsedInfo = "Completed " + targetHours + "h Goal";
                var bonusSec = elapsedSec - targetSec;
                timerDigits = formatTime(bonusSec);
                timerSub = "BONUS FASTING TIME";
                targetInfo = "Goal met at " + formatClockTime(targetEpoch);
                arcColor = 0x30D158; // Emerald
                progress = 1.0;
            } else {
                stateTitle = "FASTING " + targetHours + "/" + FastingStorage.getEatingHours();
                elapsedInfo = "Elapsed: " + elapsedH + "h " + elapsedM + "m  •  Goal: " + targetHours + "h";
                var remSec = targetSec - elapsedSec;
                timerDigits = formatTime(remSec);
                timerSub = "TIME REMAINING";
                targetInfo = "Target: " + formatClockTime(targetEpoch);
                arcColor = 0x00D2FF; // Cyan
                progress = elapsedSec.toFloat() / targetSec.toFloat();
            }
        } else if (state == FastingStorage.STATE_EATING) {
            var eatTarget = FastingStorage.getEatingHours();
            var eatTargetSec = eatTarget * 3600;
            var elapsedSec = (start > 0 && now >= start) ? (now - start) : 0;
            var elapsedH = elapsedSec / 3600;
            var elapsedM = (elapsedSec % 3600) / 60;
            var remSec = (eatTargetSec > elapsedSec) ? (eatTargetSec - elapsedSec) : 0;
            var targetEpoch = start + eatTargetSec;

            stateTitle = "EATING WINDOW";
            elapsedInfo = "Elapsed: " + elapsedH + "h " + elapsedM + "m  •  Goal: " + eatTarget + "h";
            timerDigits = formatTime(remSec);
            timerSub = "WINDOW CLOSES IN";
            targetInfo = "Closes at " + formatClockTime(targetEpoch);
            arcColor = 0xFF9F0A; // Amber
            progress = (eatTargetSec > 0) ? (elapsedSec.toFloat() / eatTargetSec.toFloat()) : 0.0;
            if (progress > 1.0) { progress = 1.0; }
        }

        // Draw track ring (dark gray)
        dc.setPenWidth(10);
        dc.setColor(0x2C2C2E, Graphics.COLOR_TRANSPARENT);
        dc.drawCircle(cx, cy, radius);

        // Draw active arc (Clockwise from top: 90 deg = 12 o'clock)
        if (progress > 0.0) {
            dc.setColor(arcColor, Graphics.COLOR_TRANSPARENT);
            var startAngle = 90;
            var sweep = (progress * 360.0).toNumber();
            if (sweep > 359) { sweep = 359; }
            var endAngle = 90 - sweep;
            dc.drawArc(cx, cy, radius, Graphics.ARC_CLOCKWISE, startAngle, endAngle);
        }

        // 1. Top State Title (e.g. FASTING 16/8)
        var yTitle = 78;
        dc.setColor(arcColor, Graphics.COLOR_TRANSPARENT);
        dc.drawText(cx, yTitle, Graphics.FONT_MEDIUM, stateTitle, Graphics.TEXT_JUSTIFY_CENTER);

        // 2. Elapsed vs Goal Info
        var yElapsed = 118;
        dc.setColor(Graphics.COLOR_LT_GRAY, Graphics.COLOR_TRANSPARENT);
        dc.drawText(cx, yElapsed, Graphics.FONT_XTINY, elapsedInfo, Graphics.TEXT_JUSTIFY_CENTER);

        // 3. Center Hero Timer Digits
        var yTimer = 168;
        dc.setColor(Graphics.COLOR_WHITE, Graphics.COLOR_TRANSPARENT);
        dc.drawText(cx, yTimer, Graphics.FONT_NUMBER_MEDIUM, timerDigits, Graphics.TEXT_JUSTIFY_CENTER);

        // 4. Timer Subtitle (Dynamic: guaranteed below digits without overlap)
        var hTimer = dc.getFontHeight(Graphics.FONT_NUMBER_MEDIUM);
        var ySub = yTimer + hTimer + 6;
        dc.setColor(0x8E8E93, Graphics.COLOR_TRANSPARENT);
        dc.drawText(cx, ySub, Graphics.FONT_XTINY, timerSub, Graphics.TEXT_JUSTIFY_CENTER);

        // 5. Target Clock Time or Status Banner
        var hSub = dc.getFontHeight(Graphics.FONT_XTINY);
        var yTarget = ySub + hSub + 26;
        if (_statusBanner != null) {
            dc.setColor(Graphics.COLOR_YELLOW, Graphics.COLOR_TRANSPARENT);
            dc.drawText(cx, yTarget, Graphics.FONT_SMALL, _statusBanner, Graphics.TEXT_JUSTIFY_CENTER);
        } else {
            dc.setColor(0xE5E5EA, Graphics.COLOR_TRANSPARENT);
            dc.drawText(cx, yTarget, Graphics.FONT_TINY, targetInfo, Graphics.TEXT_JUSTIFY_CENTER);
        }
    }

    private function formatTime(totalSec as Number) as String {
        var hours = totalSec / 3600;
        var mins = (totalSec % 3600) / 60;
        var secs = totalSec % 60;
        return hours.format("%02d") + ":" + mins.format("%02d") + ":" + secs.format("%02d");
    }

    private function formatClockTime(epochSec as Number) as String {
        if (epochSec <= 0) {
            return "--:--";
        }
        var moment = new Time.Moment(epochSec);
        var info = Gregorian.info(moment, Time.FORMAT_SHORT);
        var is24Hour = System.getDeviceSettings().is24Hour;
        var hour = info.hour;
        var ampm = "";
        if (!is24Hour) {
            ampm = (hour >= 12) ? " PM" : " AM";
            hour = hour % 12;
            if (hour == 0) { hour = 12; }
        }
        return hour.format("%d") + ":" + info.min.format("%02d") + ampm;
    }
}

import Toybox.WatchUi;
import Toybox.Graphics;
import Toybox.Time;
import Toybox.Lang;

(:glance)
class FastingGlanceView extends WatchUi.GlanceView {

    function initialize() {
        GlanceView.initialize();
    }

    function onUpdate(dc as Graphics.Dc) as Void {
        var width = dc.getWidth();
        var height = dc.getHeight();

        dc.setColor(Graphics.COLOR_BLACK, Graphics.COLOR_BLACK);
        dc.clear();

        var state = FastingStorage.getState();
        var start = FastingStorage.getStartEpoch();
        var now = Time.now().value();

        var stateText = "Fasting: Idle";
        var subText = "Press ENTER to start";
        var progressPercent = 0.0;
        var barColor = Graphics.COLOR_DK_GRAY;

        if (state == FastingStorage.STATE_FASTING) {
            var target = FastingStorage.getTargetHours();
            var elapsedSec = (start > 0 && now >= start) ? (now - start) : 0;
            var elapsedHours = elapsedSec / 3600;
            var elapsedMins = (elapsedSec % 3600) / 60;
            var targetSec = target * 3600;

            if (elapsedSec >= targetSec) {
                stateText = "Goal Reached!";
                var bonusSec = elapsedSec - targetSec;
                subText = "+" + (bonusSec / 3600) + "h " + ((bonusSec % 3600) / 60) + "m bonus";
                progressPercent = 1.0;
                barColor = 0x30D158; // Emerald
            } else {
                stateText = "Fasting: " + elapsedHours + "h " + elapsedMins + "m";
                var remSec = targetSec - elapsedSec;
                subText = (remSec / 3600) + "h " + ((remSec % 3600) / 60) + "m left";
                progressPercent = elapsedSec.toFloat() / targetSec.toFloat();
                barColor = 0x00D2FF; // Cyan
            }
        } else if (state == FastingStorage.STATE_EATING) {
            var eatTarget = FastingStorage.getEatingHours();
            var elapsedSec = (start > 0 && now >= start) ? (now - start) : 0;
            var eatTargetSec = eatTarget * 3600;
            var remSec = (eatTargetSec > elapsedSec) ? (eatTargetSec - elapsedSec) : 0;

            stateText = "Eating Window";
            subText = (remSec / 3600) + "h " + ((remSec % 3600) / 60) + "m remaining";
            progressPercent = elapsedSec.toFloat() / eatTargetSec.toFloat();
            if (progressPercent > 1.0) { progressPercent = 1.0; }
            barColor = 0xFF9F0A; // Amber
        }

        // Draw title
        dc.setColor(Graphics.COLOR_WHITE, Graphics.COLOR_TRANSPARENT);
        dc.drawText(10, 5, Graphics.FONT_GLANCE, stateText, Graphics.TEXT_JUSTIFY_LEFT);

        // Draw subtitle
        dc.setColor(Graphics.COLOR_LT_GRAY, Graphics.COLOR_TRANSPARENT);
        dc.drawText(10, 26, Graphics.FONT_GLANCE_NUMBER, subText, Graphics.TEXT_JUSTIFY_LEFT);

        // Draw mini progress bar
        var barY = height - 10;
        var barW = width - 20;
        var barH = 4;

        dc.setColor(Graphics.COLOR_DK_GRAY, Graphics.COLOR_TRANSPARENT);
        dc.fillRectangle(10, barY, barW, barH);

        if (progressPercent > 0.0) {
            var fillW = (barW * progressPercent).toNumber();
            if (fillW > barW) { fillW = barW; }
            dc.setColor(barColor, Graphics.COLOR_TRANSPARENT);
            dc.fillRectangle(10, barY, fillW, barH);
        }
    }
}

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

        try {
            var state = FastingStorage.getState();
            var start = FastingStorage.getStartEpoch();
            var now = Time.now().value();

            var stateText = "FASTING";
            var subText = "Press GPS to start";
            var progressPercent = 0.0;
            var barColor = Graphics.COLOR_DK_GRAY;

            if (state == FastingStorage.STATE_FASTING) {
                var target = FastingStorage.getTargetHours();
                var elapsedSec = (start > 0 && now >= start) ? (now - start) : 0;
                var elapsedHours = elapsedSec / 3600;
                var elapsedMins = (elapsedSec % 3600) / 60;
                var targetSec = target * 3600;

                if (elapsedSec >= targetSec) {
                    stateText = "GOAL REACHED!";
                    var bonusSec = elapsedSec - targetSec;
                    subText = "+" + (bonusSec / 3600) + "h " + ((bonusSec % 3600) / 60) + "m bonus";
                    progressPercent = 1.0;
                    barColor = 0x30D158; // Emerald
                } else {
                    stateText = "FASTING: " + elapsedHours + "h " + elapsedMins + "m";
                    var remSec = targetSec - elapsedSec;
                    subText = (remSec / 3600) + "h " + ((remSec % 3600) / 60) + "m left • Goal " + target + "h";
                    progressPercent = elapsedSec.toFloat() / targetSec.toFloat();
                    barColor = 0x00D2FF; // Cyan
                }
            } else if (state == FastingStorage.STATE_EATING) {
                var eatTarget = FastingStorage.getEatingHours();
                var elapsedSec = (start > 0 && now >= start) ? (now - start) : 0;
                var eatTargetSec = eatTarget * 3600;
                var remSec = (eatTargetSec > elapsedSec) ? (eatTargetSec - elapsedSec) : 0;

                stateText = "EATING WINDOW";
                subText = (remSec / 3600) + "h " + ((remSec % 3600) / 60) + "m left • Goal " + eatTarget + "h";
                progressPercent = (eatTargetSec > 0) ? (elapsedSec.toFloat() / eatTargetSec.toFloat()) : 0.0;
                if (progressPercent > 1.0) { progressPercent = 1.0; }
                barColor = 0xFF9F0A; // Amber
            }

            // 1. Draw Title (e.g. FASTING: 14h 22m)
            dc.setColor(Graphics.COLOR_WHITE, Graphics.COLOR_TRANSPARENT);
            dc.drawText(12, 12, Graphics.FONT_GLANCE, stateText, Graphics.TEXT_JUSTIFY_LEFT);

            // 2. Draw Subtitle (e.g. 1h 38m left • Goal 16h)
            dc.setColor(Graphics.COLOR_LT_GRAY, Graphics.COLOR_TRANSPARENT);
            dc.drawText(12, 42, Graphics.FONT_XTINY, subText, Graphics.TEXT_JUSTIFY_LEFT);

            // 3. Draw Progress Bar
            var barX = 12;
            var barY = 72;
            var barW = width - 24;
            var barH = 6;

            dc.setColor(0x2C2C2E, Graphics.COLOR_TRANSPARENT);
            dc.fillRectangle(barX, barY, barW, barH);

            if (progressPercent > 0.0) {
                var fillW = (barW * progressPercent).toNumber();
                if (fillW > barW) { fillW = barW; }
                if (fillW < 4) { fillW = 4; }
                dc.setColor(barColor, Graphics.COLOR_TRANSPARENT);
                dc.fillRectangle(barX, barY, fillW, barH);
            }
        } catch (e) {
            // Emergency fallback if storage fails
            dc.setColor(Graphics.COLOR_WHITE, Graphics.COLOR_TRANSPARENT);
            dc.drawText(12, 20, Graphics.FONT_GLANCE, "FASTING", Graphics.TEXT_JUSTIFY_LEFT);
            dc.setColor(Graphics.COLOR_LT_GRAY, Graphics.COLOR_TRANSPARENT);
            dc.drawText(12, 48, Graphics.FONT_XTINY, "Press GPS to open", Graphics.TEXT_JUSTIFY_LEFT);
        }
    }
}

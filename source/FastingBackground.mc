import Toybox.System;
import Toybox.Background;

(:background)
class FastingServiceDelegate extends System.ServiceDelegate {

    function initialize() {
        ServiceDelegate.initialize();
    }

    function onTemporalEvent() as Void {
        try {
            Background.requestApplicationWake("Fasting window complete!");
        } catch (e) {
            System.println("Background wake error: " + e.getErrorMessage());
        }
        Background.exit(true);
    }
}

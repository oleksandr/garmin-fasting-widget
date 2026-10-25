import Toybox.Communications;
import Toybox.Application.Properties;
import Toybox.System;
import Toybox.Time;
import Toybox.Lang;

class FastingSync {

    private var _callback as Method(success as Boolean, message as String) as Void;
    private var _pendingIds as Array<Number>;

    function initialize(callback as Method(success as Boolean, message as String) as Void) {
        _callback = callback;
        _pendingIds = [];
    }

    function startSync() as Void {
        var unsynced = FastingStorage.getUnsyncedFasts();
        if (unsynced.size() == 0) {
            _callback.invoke(true, "All synced");
            return;
        }

        var url = "https://fasting.home.lobunets.me/api";
        var token = "your_secret_token_here";

        try {
            var urlProp = Properties.getValue("apiUrl");
            if (urlProp != null && urlProp instanceof String && (urlProp as String).length() > 0) {
                url = urlProp as String;
            }
            var tokenProp = Properties.getValue("apiToken");
            if (tokenProp != null && tokenProp instanceof String) {
                token = tokenProp as String;
            }
        } catch (e) {
            System.println("Property load error: " + e.getErrorMessage());
        }

        _pendingIds = [];
        for (var i = 0; i < unsynced.size(); i++) {
            var item = unsynced[i];
            var id = item.get("id");
            if (id != null) {
                _pendingIds.add(id as Number);
            }
        }

        var payload = {
            "device" => "Instinct 3 AMOLED 50mm",
            "synced_at" => Time.now().value(),
            "fasts" => unsynced
        };

        var options = {
            :method => Communications.HTTP_REQUEST_METHOD_POST,
            :headers => {
                "Content-Type" => Communications.REQUEST_CONTENT_TYPE_JSON,
                "Authorization" => "Bearer " + token
            },
            :responseType => Communications.HTTP_RESPONSE_CONTENT_TYPE_JSON
        };

        Communications.makeWebRequest(url, payload, options, method(:onReceive));
    }

    function onReceive(responseCode as Number, data as Dictionary or String or Null) as Void {
        if (responseCode == 200) {
            if (_pendingIds.size() > 0) {
                FastingStorage.markFastsSynced(_pendingIds);
            }
            _callback.invoke(true, "Synced " + _pendingIds.size() + " items");
        } else {
            var err = "Err " + responseCode;
            if (responseCode == -400) {
                err = "Invalid URL / Cert";
            } else if (responseCode == -1001) {
                err = "SSL / Net Error";
            } else if (responseCode == -2) {
                err = "No BLE / Phone";
            }
            _callback.invoke(false, err);
        }
    }
}

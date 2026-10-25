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
        if (!System.getDeviceSettings().phoneConnected) {
            _callback.invoke(false, "Phone Disconnected");
            return;
        }

        var unsynced = FastingStorage.getUnsyncedFasts();

        // If currently in an active fast, include the active session in the sync payload
        if (FastingStorage.getState() == FastingStorage.STATE_FASTING && FastingStorage.getStartEpoch() > 0) {
            var start = FastingStorage.getStartEpoch();
            var now = Time.now().value();
            var activeRecord = {
                "id" => start,
                "start" => start,
                "end" => 0,
                "target_hours" => FastingStorage.getTargetHours(),
                "duration_hours" => (now - start) / 3600.0,
                "state" => "FASTING",
                "synced" => false
            };
            unsynced.add(activeRecord);
        }

        if (unsynced.size() == 0) {
            _callback.invoke(true, "All synced");
            return;
        }

        var url = "https://fasting.lobunets.home.network/api";
        var token = "";

        try {
            var urlProp = Properties.getValue("apiUrl");
            if (urlProp != null && urlProp instanceof String) {
                var u = urlProp as String;
                // If it contains the obsolete lan.home dummy host, reset and ignore it
                if (u.find("fasting.lan.home") != -1 || u.length() == 0) {
                    Properties.setValue("apiUrl", url);
                } else {
                    url = u;
                }
            } else {
                Properties.setValue("apiUrl", url);
            }

            var tokenProp = Properties.getValue("apiToken");
            if (tokenProp != null && tokenProp instanceof String) {
                var t = tokenProp as String;
                if (t != "your_secret_token_here") {
                    token = t;
                }
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

        var headers = {
            "Content-Type" => Communications.REQUEST_CONTENT_TYPE_JSON
        };
        if (token.length() > 0) {
            headers.put("Authorization", "Bearer " + token);
        }

        var options = {
            :method => Communications.HTTP_REQUEST_METHOD_POST,
            :headers => headers,
            :responseType => Communications.HTTP_RESPONSE_CONTENT_TYPE_JSON
        };

        Communications.makeWebRequest(url, payload, options, method(:onReceive));
    }

    function onReceive(responseCode as Number, data as Dictionary or String or Null) as Void {
        if (responseCode == 200) {
            if (_pendingIds.size() > 0) {
                FastingStorage.markFastsSynced(_pendingIds);
            }
            _callback.invoke(true, "Synced OK");
        } else {
            var err = "Err " + responseCode;
            if (responseCode == -104 || responseCode == -2) {
                err = "Phone BLE Offline";
            } else if (responseCode == -300) {
                err = "Err -300 (Timeout)";
            } else if (responseCode == -400) {
                err = "Invalid URL";
            } else if (responseCode == -1001) {
                err = "SSL / Net Error";
            } else if (responseCode == -2) {
                err = "No BLE / Phone";
            }
            _callback.invoke(false, err);
        }
    }
}

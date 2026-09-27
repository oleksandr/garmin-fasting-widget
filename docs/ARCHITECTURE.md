# Garmin Fasting Widget - Architecture & Design Specification

## 1. System Overview

The **Garmin Fasting Widget** is an intermittent fasting companion built specifically for the **Garmin Instinct 3 AMOLED (50mm)** using the official Garmin Connect IQ SDK and the Monkey C programming language.

It is designed as an interactive **Glance Widget** rather than a Watch Face:
1. **Interactive State Management**: Allows starting, stopping, extending, and shifting fasting/eating windows using the watch's 5 physical buttons (Instinct 3 has no touchscreen).
2. **True Low-Power Design**: Consumes **0% idle battery** while on the watch face (unlike custom watch faces that run continuously 24/7).
3. **Wrist Vibration / Alerts**: Leverages Connect IQ's `Background.registerForTemporalEvent` and `Background.requestApplicationWake` to trigger haptic system notifications when target fasting windows conclude.
4. **Local Network Sync**: Stores a rolling history of the last 30 fasts locally and syncs them to a private home-network API via Bluetooth proxy through Garmin Connect Mobile.

---

## 2. State Machine

The widget operates as a deterministic finite state machine (FSM):

```
       ┌────────────────────────────────────────────────────────┐
       │                                                        │
       ▼                                                        │
┌──────────────┐         Start Fast (ENTER)            ┌─────────────────┐
│     IDLE     │ ────────────────────────────────────► │     FASTING     │
└──────────────┘                                       └─────────────────┘
       ▲                                                        │
       │                                                        │ Time >= Target (e.g. 16h)
       │                                                        ▼
       │                                               ┌─────────────────┐
       │                                               │  GOAL REACHED   │
       │                                               │ (Bonus Counter) │
       │                                               └─────────────────┘
       │                                                        │
       │ Break Fast / Start Eating (ENTER)                      │ Start Eating (ENTER)
       │                                                        ▼
       │                                               ┌─────────────────┐
       └────────────────────────────────────────────── │  EATING WINDOW  │
                                                       │   (8h Window)   │
                                                       └─────────────────┘
```

### State Definitions
1. **IDLE**: No active fast or eating window running. Ready for start.
2. **FASTING**: Active fast counting up toward target hours (default 16h).
   - Display: Cyan countdown arc, elapsed time, hours remaining.
   - Background event: Scheduled for `start_epoch + (target_hours * 3600)`.
3. **GOAL REACHED**: Fasting goal reached (e.g. exceeded 16 hours).
   - Display: Green/Gold celebratory arc, counting "Bonus Fasting Time" until the user explicitly breaks the fast.
4. **EATING WINDOW**: Active eating window counting down (default 8h).
   - Display: Warm Amber countdown arc, reminding user when the eating window closes.
   - Button actions: Allows shifting eating window or immediately starting the next fast.

---

## 3. Data Model & Persistent Storage

Stored in `Toybox.Application.Storage`:

### Active State Keys
* `currentState` (Number): Enum `0=IDLE, 1=FASTING, 2=GOAL_REACHED, 3=EATING`
* `activeStartEpoch` (Number): Unix timestamp in seconds when current phase started
* `targetHours` (Number): Desired fasting duration (e.g., 16)
* `eatingWindowHours` (Number): Desired eating window (e.g., 8)
* `pendingWakeEpoch` (Number): Scheduled wake-up timestamp

### Fast Record Schema (History Buffer)
A rolling FIFO array `fastingHistory` capped at **30 records**:
```json
[
  {
    "id": 1727434800,
    "start": 1727434800,
    "end": 1727492400,
    "target": 16,
    "duration": 16.0,
    "state": "COMPLETED",
    "synced": false
  }
]
```

---

## 4. Garmin Platform Capabilities & Constraints

| Feature | Connect IQ Capability | Implementation in this App |
| :--- | :--- | :--- |
| **Glance Loop** | `AppBase.getGlanceView()` | Compact view in the UP/DOWN widget list showing fast progress without opening the app. |
| **Vibration** | `Toybox.Attention.vibrate()` | Used inside the widget when pressing buttons or opening confirmation dialogs. |
| **Background Alerts** | `Background.requestApplicationWake()` | Schedules a temporal event at `start + target`. When reached, the OS wakes and presents a system haptic prompt to the user. |
| **Display Optimization** | 390 × 390 AMOLED | Background is set to pure black (`#000000`) so inactive OLED pixels consume 0 power. |
| **Button Navigation** | `BehaviorDelegate` | 100% button-driven (Instinct 3 has no touchscreen). |

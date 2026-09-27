# UI Specification: Instinct 3 AMOLED 50mm (390x390)

## 1. Display & Hardware Properties

* **Resolution**: 390 × 390 pixels (Full Circular display).
* **Display Technology**: AMOLED (High-contrast, self-lit pixels).
* **Color Depth**: 24-bit / 65K color palette.
* **Input**: 5 physical buttons (NO touchscreen).
  * **Top Right (`GPS / ENTER`)**: Confirm / Select / Open Actions Menu.
  * **Bottom Right (`SET / BACK`)**: Back / Dismiss / Exit.
  * **Middle Left (`UP / MENU`)**: Quick Shift Window (+15 min) or Long Press for Settings.
  * **Bottom Left (`DOWN`)**: Quick Shift Window (-15 min).
  * **Top Left (`LIGHT`)**: Standard system backlight/power.

---

## 2. AMOLED Color Palette & Power Efficiency

To maximize battery life on the AMOLED panel, the background must always be **pure black (`#000000`)**, ensuring inactive pixels consume zero power.

| Token | Hex Code | Visual Meaning |
| :--- | :--- | :--- |
| `COLOR_BG` | `#000000` | True Black (0 power draw) |
| `COLOR_FASTING` | `#00D2FF` | Cool Cyan: Active Fasting phase |
| `COLOR_GOAL` | `#30D158` | Vibrant Emerald/Gold: Goal Achieved (>16h) |
| `COLOR_EATING` | `#FF9F0A` | Warm Amber: Active Eating Window |
| `COLOR_TEXT_MUTED` | `#8E8E93` | Neutral Gray: Sub-labels & units |
| `COLOR_TEXT_BRIGHT` | `#FFFFFF` | Crisp White: Primary timer digits |

---

## 3. Screen Layouts

### Layout A: Glance View (UP/DOWN Widget Loop)
When scrolling through your Garmin widget list, this view renders within the glance slot:

```
┌──────────────────────────────────────────────┐
│ [FAST] 14h 22m / 16h        (1h 38m left)    │
│ [████████████████████░░░░░] 89%              │
└──────────────────────────────────────────────┘
```
* **Text Line 1**: Current state tag `[FAST]` or `[EAT]`, elapsed time, target.
* **Text Line 2 / Bar**: Progress bar spanning width, showing percentage completion.

### Layout B: Full Main View (390 × 390 Round AMOLED)
Accessed by pressing `ENTER` on the Glance:

```
                   ▲ [UP: +15m]
                 .  - - -  .
             . '             ' .
           /    [FASTING 16/8]   \
          /                       \
         |     Elapsed: 14h 22m    |
         |                         |
         |      01:37:45           |  ◄ Big Timer: Time Remaining
         |       REMAINING         |
         |                         |
         |     Target: 12:30 PM    |
          \                       /
           \   [Press GPS: MENU] /
             ' .             . '
                 ' - - - - '
                   ▼ [DOWN: -15m]
```

* **Outer Ring (`dc.drawArc`)**:
  * Outer Radius: `185px`, Stroke: `10px`.
  * Arc starts at 12 o'clock (`90°`) and advances clockwise up to `360°` based on percentage completed.
  * Cyan during fasting; shifts to Emerald Green once `100%` is passed.
* **Header**: Small font, indicates state (`FASTING 16:8` or `EATING 8h`).
* **Center**: Large numeric font displaying remaining or elapsed time.
* **Footer**: Action hint (`GPS: ACTIONS`).

### Layout C: Action Menu (Triggered by `ENTER`)
Pressing `ENTER` opens a native Garmin menu:
1. **End Fast & Start Eating**: Logs current fast, switches state to `EATING`, schedules eating window alarm.
2. **Shift Start Time**: Sub-menu to add/subtract 15m or 1h (if dinner was earlier/later than logged).
3. **Sync Now**: Attempts immediate HTTP push of unsynced records to local home API.
4. **History**: View summary of last 7 fasts and average duration.
5. **Reset / Cancel**: Cancel active session.

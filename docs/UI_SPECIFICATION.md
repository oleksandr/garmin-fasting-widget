# UI Specification: Instinct 3 AMOLED 50mm (416 × 416)

## 1. Display & Hardware Properties

* **Resolution**: 416 × 416 pixels (Full Circular display, center `cx=208, cy=208`).
* **Display Technology**: AMOLED (High-contrast, self-lit pixels).
* **Color Depth**: 24-bit / 65K color palette.
* **Input**: 5 physical buttons (NO touchscreen).
  * **Top Right (`GPS / ENTER`)**: Confirm / Select / Open Actions Menu.
  * **Bottom Right (`SET / BACK`)**: Back / Dismiss / Exit.
  * **Middle Left (`UP / MENU`)**: Quick Shift Window (+15 min).
  * **Bottom Left (`DOWN`)**: Quick Shift Window (-15 min).
  * **Top Left (`LIGHT`)**: Standard system backlight/power.

---

## 2. AMOLED Color Palette & Power Efficiency

To maximize battery life on the AMOLED panel, the background must always be **pure black (`#000000`)**, ensuring inactive pixels consume zero power.

| Token | Hex Code | Visual Meaning |
| :--- | :--- | :--- |
| `COLOR_BG` | `#000000` | True Black (0 power draw) |
| `COLOR_FASTING` | `#00D2FF` | Cool Cyan: Active Fasting phase |
| `COLOR_GOAL` | `#30D158` | Vibrant Emerald: Goal Achieved (>16h) |
| `COLOR_EATING` | `#FF9F0A` | Warm Amber: Active Eating Window |
| `COLOR_TEXT_MUTED` | `#8E8E93` | Neutral Gray: Sub-labels & units |
| `COLOR_TEXT_BRIGHT` | `#FFFFFF` | Crisp White: Primary timer digits |

---

## 3. Screen Layouts

### Layout A: Glance View (UP/DOWN Widget Loop)
When scrolling through the widget list, this view renders within the glance slot:

```
┌──────────────────────────────────────────────┐
│ FASTING: 14h 22m                             │
│ 1h 38m left • Goal 16h                       │
│ [████████████████████░░░░░]                  │
└──────────────────────────────────────────────┘
```
* **Text Line 1**: Active state and elapsed time.
* **Text Line 2**: Remaining time and goal duration.
* **Text Line 3**: Horizontal progress bar.

### Layout B: Main View (416 × 416 Round AMOLED)
Accessed by pressing `ENTER` on the Glance:

```
                     .  - - -  .
                 . '     ▲       ' .
               /    [FASTING 16/8]   \
              /                       \
             |     Elapsed: 14h 22m    |
             |                         |
             |        01:37:45         |  ◄ Hero Timer
             |     TIME REMAINING      |
             |                         |
             |     Target: 8:00 PM     |
              \                       /
               \         ▼           /
                 ' .             . '
                     ' - - - - '
```

* **Outer Track & Progress Arc**:
  * Outer Radius: `194px`, Stroke: `10px`.
  * Starts at 12 o'clock (`90°`) and advances clockwise up to `360°`.
  * Cyan during fasting; shifts to Emerald Green once goal is reached.
* **Header**: State title (`FASTING 16/8`, `GOAL REACHED!`, or `EATING WINDOW`).
* **Elapsed Info**: Elapsed duration and current goal.
* **Hero Timer**: Large numeric font (`FONT_NUMBER_MEDIUM`) with `TIME REMAINING` label underneath.
* **Footer**: Target completion clock time (`Target: 8:00 PM`) or temporary feedback banner (`+15 min added`).

### Layout C: Action Menu (Triggered by `ENTER`)
1. **End Fast & Start Eating**: Logs current fast, switches state to `EATING`.
2. **Shift Start Time**: Add/subtract 15m or 1h.
3. **Sync Now**: Immediate HTTP push of unsynced records to local home API.
4. **History**: Summary of last 30 fasts.
5. **Reset / Cancel**: Cancel active session.

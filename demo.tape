# demo.tape
Output demo.gif

Set FontSize 14
Set Width 1920
Set Height 1080
Set Shell "bash"
Set TypingSpeed 250ms

Hide
Type "export PATH=$PATH:/home/asyncerror/developer/poker-holdem-ranges-tui/tmp && cd /home/asyncerror/developer/poker-holdem-ranges-tui && clear"
Enter
Sleep 1s
Show

Type "main examples/sideranges/"
Enter
Sleep 3s

# Grid cursor navigation
Type "l"
Sleep 1s
Type "l"
Sleep 1s
Type "j"
Sleep 1s
Type "j"
Sleep 3s

# Switch to 25BB tab (has sideranges)
Tab
Sleep 3s

# Enter sideranges with 's'
Type "s"
Sleep 2s

# Navigate to "vs BTN"
Type "j"
Sleep 2s

# Load siderange with Enter
Enter
Sleep 3s

# Navigate grid in loaded siderange
Type "l"
Sleep 1s
Type "j"
Sleep 1s
Type "j"
Sleep 3s

# Go back to sideranges and switch to "vs CO"
Type "s"
Sleep 1s
Type "k"
Sleep 1s
Enter
Sleep 3s

# Navigate to mixed hand QQ
Type "k"
Sleep 500ms
Type "k"
Sleep 500ms
Type "h"
Sleep 500ms
Type "h"
Sleep 3s

# Esc to restore original range
Escape
Sleep 3s

Type "q"

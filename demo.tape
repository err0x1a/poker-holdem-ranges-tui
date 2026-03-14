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

Type "main examples/"
Enter
Sleep 3s

# Navigate list to tab range file
Ctrl+N
Sleep 2s
Ctrl+N
Sleep 2s
Ctrl+N
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

# Switch tab to show mixed hands with border progression
Tab
Sleep 3s

# Navigate to mixed hand
Type "k"
Sleep 1s
Type "l"
Sleep 3s

# Go to CO vs BTN 3-bet example (lots of mixed hands)
Ctrl+N
Sleep 2s
Ctrl+N
Sleep 3s

# Navigate to QQ (mixed 4-bet/call)
Type "h"
Sleep 500ms
Type "h"
Sleep 500ms
Type "h"
Sleep 500ms
Type "k"
Sleep 500ms
Type "k"
Sleep 3s

# Navigate to AKo (mixed 4-bet/call)
Type "l"
Sleep 500ms
Type "l"
Sleep 500ms
Type "l"
Sleep 3s

# Navigate to ATs (mixed call/bluff)
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 3s

# Toggle opposite range
Ctrl+O
Sleep 4s

# Back to original
Ctrl+O
Sleep 2s

Type "q"

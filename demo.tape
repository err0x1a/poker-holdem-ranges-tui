# demo.tape
Output demo.gif

Set FontSize 14
Set Width 1920
Set Height 1080
Set Shell "bash"
Set TypingSpeed 250ms

Hide
Type "export PATH=$PATH:/home/asyncerror/developer/poker-holdem-ranges-tui && cd /home/asyncerror/developer/poker-holdem-ranges-tui && clear"
Enter
Sleep 1s
Show

Type "phr-tui examples/"
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

# Switch tab
Tab
Sleep 3s

# Navigate to mixed hand
Type "k"
Sleep 1s
Type "l"
Sleep 3s

# Navigate to opposite range example
Ctrl+N
Sleep 3s

# Toggle opposite range
Type "o"
Sleep 4s

# Back to original
Type "o"
Sleep 2s

Type "q"

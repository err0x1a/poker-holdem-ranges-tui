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

# Navigate to mixed hand to show border progression
Type "k"
Sleep 1s
Type "l"
Sleep 3s

# Navigate to opposite range example
Ctrl+N
Sleep 3s

# Toggle opposite range (Ctrl+O)
Ctrl+O
Sleep 4s

# Back to original
Ctrl+O
Sleep 2s

Type "q"

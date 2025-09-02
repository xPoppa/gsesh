## Thoughts
* The session manager should be a daemon running in the background (systemctl?)
* Have to make a cli client probably next to it
* Session manager never stops running except if you explictly kill it
* How much state do I want to save? 
    * For now not so much. Just the pid so you can switch easily
* Am I just creating my own tmux? LOL
    * Kind of but then just with ghostty :D

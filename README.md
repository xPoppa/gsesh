## Thoughts

* The session manager should be a daemon running in the background (systemctl?)
    * <https://ieftimov.com/posts/four-steps-daemonize-your-golang-programs/>
    * Session manager never stops running except if you explictly kill it

* Have to make a cli client probably next to it that sends command to the daemon 
    * Which protocol?
    * <https://github.com/nikzayn/golang-system-design/blob/main/client_server/server/main.go>
    * <https://github.com/ChrIgiSta/unix_sockets>

* How much state do I want to save? 
    * For now not so much. Just the pid so you can switch easily

* Am I just creating my own tmux? LOL
    * Kind of but then just with ghostty :D

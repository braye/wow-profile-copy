# wow-profile-copy

This TUI utility provides an easy way to copy addon settings, keybinds, and macros between characters, or even different versions of the WoW client (e.g. PTR).

It does not currently copy client settings (graphics, sound levels, etc) between versions of the game.

# FAQ

## My keybinds aren't copying correctly!

Disable keybind synchronization.

When logged in, put this in the chat window:
```
/console synchronizeBindings 0
```

Close the game, Re-run the sync.

Open the game, and run:
```
/console synchronizeBindings 1
```
Change *any* key binding, and apply the changes - you can change it back later.

This "tricks" the WoW client into accepting the new keybindings, and saving them to Blizzard's servers. Otherwise, it sees that the keybindings for the account don't match the ones saved on the server, and "helpfully" changes them.

Leaving `synchronizeBindings` turned off entirely also solves the issue.
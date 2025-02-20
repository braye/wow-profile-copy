# wow-profile-copy

## This software is feature-complete.

Development is continuing on [wow-profile-copy-ng](https://github.com/braye/wow-profile-copy-ng), which adds a GUI.

This TUI utility provides an easy way to copy addon settings, keybinds, and macros between characters, or even different versions of the WoW client (e.g. PTR).

It does not currently copy client settings (graphics, sound levels, etc) between versions of the game.

## Backups

This software overwrites large amounts of configuration data automatically. There is no "undo". For that reason, before you copy profiles around, you should consider taking a backup.

To do that, simply copy the entire `WTF` folder in the version folder that you're going to be copying to (e.g. `_classic_ptr_`) to somewhere safe before running the tool.

# FAQ

## My keybinds aren't copying correctly!

Disable keybind synchronization.

Log into the game version that configs are being copied to.

Type this in the chat window:
```
/console synchronizeBindings 0
```

Close the game, Re-run the sync.

Open the game, and run:
```
/console synchronizeBindings 1
```
In the chat window.

Change *any* key binding, and apply the changes - you can change it back later.

This "tricks" the WoW client into accepting the new keybindings, and saving them to Blizzard's servers. Otherwise, it sees that the keybindings for the account don't match the ones saved on the server, and "helpfully" changes them.

Leaving `synchronizeBindings` turned off entirely also solves the issue.

## How do I copy the spell placements on my hotbars?

As far as I can tell, spell -> bar slot assignments are saved to the realm. That means there's no way for this tool to copy them, because the data doesn't exist on your computer. However, something like [MySlot](https://github.com/tg123/myslot) can help with that.

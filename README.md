# enfolder

enfolder helps you making your folders (like download folder) tidy by moving files with certain keywords (case insensitively) into desired folder.

With following config file: `enfolder_rule.json`, enfolder will move all files that filename contains "vlcsnap" into the folder "vlcsnap",
Note that the folder with exact the same name to the destination folder will not be moved.

```json
[
  {
    "folder_name": "vlcsnap",
    "key_words": [
      "vlcsnap"
    ]
  }
]
```

# bubblesgum - bubbles general utility manager

a few ways to compose multiple [bubbles](https://github.com/charmbracelet/bubbles) into one layout.

## boxer 🥊 - compose bubbles into boxes 📦

To layout the bubbles with boxer, one would construct a layout-tree 🌲.
Each node holds a arbitraly amout of children as well as the orientation of them (Horizonal/Vertical).
The Leaves are responsible for the focus, the box arround the content and hold the content(-bubbles) it self.

```
╭l1────────────────────╮╭l2────────╮╭l3────────╮
│ 1╭>list of something ││ some   0 ││ a        │               V
│  │ ----------------- ││ status 1 ││  text    │              / \
│ 2├ which you may     ││ infor- 2 ││   logo   │             /   \
│  │ edit as you wish  ││ mation 4 ││    even! │            H    l5
│ 3├ or just use it    │╰──────────╯╰──────────╯           / \
│ 4├ to display        │╭l4────────────────────╮          /   \
│ 5├ and scroll        ││ Maybe here is a      │         l1    V
│                      ││ note written to each │              / \
│                      ││ list entry, in a     │             /   \
│                      ││ bubbles viewport.    │            H    l4
│                      ││                      │           / \
╰──────────────────────╯╰──────────────────────╯          /   \
╭l5────────────────────────────────────────────╮         l2   l3
│ maybe a progressbar or a command input? 100% │
╰──────────────────────────────────────────────╯
```

## list - compose bubbles into a scrollable list 📜

Simply to list multiline textitems or other bubbles like textinputs,
with custom pre- or suffix to convey the linenumber,
the current item or something of your own.

## LICENSE

MIT

package main

type asciiFont struct {
    numbers map[int]string
    background string
}

var pagga asciiFont = asciiFont{
    numbers: map[int]string{
        0:
"░░░\n"+
"░░░\n"+
"░░░",
        1:
"▀█░\n"+
"░█░\n"+
"▀▀▀",
        2:
"▀▀▄\n"+
"▄▀░\n"+
"▀▀▀",
        3:
"▀▀█\n"+
"░▀▄\n"+
"▀▀░",
        4:
"█░█\n"+
"░▀█\n"+
"░░▀",
        5:
"█▀▀\n"+
"▀▀▄\n"+
"▀▀░",
        6:
"▄▀▀\n"+
"█▀▄\n"+
"░▀░",
        7:
"▀▀█\n"+
"▄▀░\n"+
"▀░░",
        8:
"▄▀▄\n"+
"▄▀▄\n"+
"░▀░",
        9:
"▄▀▄\n"+
"░▀█\n"+
"▀▀░",
},
    background: "░",
}


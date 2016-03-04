# goboy

Game Boy emulator written in Go. I've always wanted to write an emulator.

## Compiling

Dependencies:

```sh
go get github.com/deweerdt/gocui
```

Run:

```sh
go run *.go <data-file>
```

## Resources

- http://marc.rawer.de/Gameboy/Docs/GBCPUman.pdf
- http://www.z80.info/z80code.htm
- http://imrannazar.com/Gameboy-Z80-Opcode-Map
- http://gameboy.mongenel.com/dmg/opcodes.html
- http://www.pastraiser.com/cpu/gameboy/gameboy_opcodes.html

### DAA

**Resources:**

- http://imrannazar.com/Binary-Coded-Decimal-Addition-on-Atmel-AVR
- http://stackoverflow.com/a/29990058

Tested in Scala:

```scala
scala> def daa(t: (Boolean, Boolean, Boolean, Int)) = { var (fn, fc, fh, r) = t; val i = if (fn) -1 else 1; if (fh || (r&0xf) > 0x09) r += (i * 0x06); if (fc || r > 0x99) (true, r + (i * 0x60)) else (false, r) }
daa: (t: (Boolean, Boolean, Boolean, Int))(Boolean, Int)

scala> def add(a : Int, b: Int) = { val r = a+b; val fh = ((a^b^r)&0x10) > 0; val fc = r > 0xff; (false, fc, fh, r&0xff) }
add: (a: Int, b: Int)(Boolean, Boolean, Boolean, Int)

scala> (for { a <- 0 to 99; b <- 0 to 99; (fc, r) = daa(add(toBcd(a), toBcd(b))) } yield (toBcd((a+b)%100), r%256)) forall (t => t._1 == t._2)
res49: Boolean = true

scala> (for { a <- 0 to 99; b <- 0 to 99; (fc, r) = daa(add(toBcd(a), toBcd(b))) } yield (fc, a + b >= 100)) forall (t => t._1 == t._2)
res51: Boolean = true

scala> def sub(a: Int, b: Int) = { var r = a-b; val fh = ((a^b^r)&0x10) > 0; val fc = r < 0; (true, fc, fh, r&0xff) }
sub: (a: Int, b: Int)(Boolean, Boolean, Boolean, Int)

scala> (for { a <- 0 to 99; b <- 0 to 99; (fc, r) = daa(sub(toBcd(a), toBcd(b))) } yield (toBcd(((a-b)%100+100)%100), r%256)) forall (t => t._1 == t._2)
res57: Boolean = true

scala> (for { a <- 0 to 99; b <- 0 to 99; (fc, r) = daa(sub(toBcd(a), toBcd(b))) } yield (fc, a < b)) forall (t => t._1 == t._2)
res59: Boolean = true
```

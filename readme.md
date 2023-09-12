
可以在通过一个数组来实现
	按照Packet距离，而不是group距离，来进行丢包判别与恢复。




































































TODO: 可以通过组间交叉抵抗连续丢包

​	使用交叉组是因为异或无法做到RS编码参数的自由。

​	组间交叉无用，不如一组多个冗余包；多个冗余包不会增加冗余度参数个数，只是可以抵抗连续丢包的能力。



工作原理：

​	当冗余度小于1时，	将N个数据包划分为一个组，每个组中有一个冗余包；当冗余度大于等于1时，发送重复数据包。

​	冗余包是所有数据包Byte异或的值，恢复时，只需要将N-1个数据包和冗余包进行异或，即可得到丢失的数据包。

​	无论发送还是接收，都不需要暂存所有的数据；发送、接收时都只需要维护一个数据包大小的内存即可，接收时不要求数据包有序、恢复时不需要知道丢失的是哪个数据包。







change: 
	parityBlocks可以大于1，大于1时将进行组间交叉。


TODO: 解决尾0问题
		1. 封包解决, xor包括（部分、计划把Group-Len两位出来拿来，这样将导致Parity包的Group-Len失效）包头，包头flag用两位表示尾0替换


```go
	//	redundancy = parityBlocks / (dataBlocks + parityBlocks)
	if redundancy <= 0 || redundancy >= 1 {
		panic(errors.New("redundancy must be in (0,1)"))
	}
	x := int(redundancy * 1e4)
	y := 1e4 - x
	{ // greatest common divisor
		var gcd int
		for {
			gcd = (x % y)
			if gcd > 0 {
				x = y
				y = gcd
			} else {
				gcd = y
				break
			}
		}
		x, y = x/gcd, y/gcd
	}
	for x > 255 || y > 255 {
		factor := int(math.Ceil(float64(max(x, y)) / 256))
		f.dataBlocks = max(uint8(y/factor), 1)
		f.parityBlocks = max(uint8(x/factor), 1)
	}
```
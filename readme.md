

# A-FEC

Adaptive Forward Error Correction, 自适应前向纠错。

**本版本不支持组间交叉，支持0冗余比，支持大于1的冗余比**



###### 工作原理：

- 自适应

  在双工链路上，通过数据包包头携带丢包信息，实现纠错能力的自适应。

- 前向纠错恢复：

  将N个数据包（DataBlock）划分为不同的组，每个组对应一个ParityBlock，ParityBlock是一个组中所有DataBlock异或的结果，它在组的最后才会被发送，因此一个组最多能恢复一个丢包的。通过调整组中数据包个N来改变改变链路的数据冗余度、纠错能力。当冗余比为0是，N=1、且不发送ParityBlock。

  > 每个组只有一个ParityBlock，只能适应冗余比小于1的情况；改进参考 组间交叉

​				

###### A-FEC 格式：

所有的数据包（包括ParityBlock）都遵循以下格式（Upack）：

```json
	Upack{
		Payload: {
			data(nB)
		}

		Tail-Header: {
			Zeros-Fill(2b)    : 尾0填充
			Group-Len(8b)     : 一个组中DataBlock个数
			Group-Idx(1B)     : 组ID
			Lossy-Perc(1B)    : 对端链路丢包率
			Group-Flag(1B)    : 组标志
		}
	}
```

关于尾0填充：由于根据异或运算实现恢复，如果被恢复数据包不是组中最大的数据包，那么尾部将会是0；如果原始数据以0结尾，这将无法区分。尾0填充方法：获取原始数据中最后一个非0值，然后从1或2中选取一个与其不相等的值作为填充值，将尾0替换为此值，并甚至Zeros-Fill；如果原始数据全是0，填充值为1。



###### 组间交叉：

组间交叉可以改进抗连续丢包的能力，不能实现颗粒度更细的参数控制。





如果当前要求冗余比为1.4，那么组的大小为5，每个DataBlock重复发送一次, 并且这个组







































































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
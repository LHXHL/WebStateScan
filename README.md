# WebStateScan

**基于golang写的一个小工具，主要用于探测目标的状态。**

## **用法**

在input.txt中放入所需要探测的URL,支持test.com,http://test.com,ip:port各种形式

```go
git clone https://github.com/LHXHL/WebStateScan.git
go build -o WebStateScan
./WebStateScan -h
```

![image-20230511120011456](https://p.ipic.vip/ctr7uk.png)

**彩色输出结果(-m可以静默输出，直接在文件中查看结果)**

![image-20230511115748544](https://p.ipic.vip/un6n4v.png)

**按照状态码从小到大的顺序排列，方便查看。**

![image-20230511120132346](https://p.ipic.vip/vva6k1.png)

## ToDo...

**~~改为命令flag模式的输入~~**

~~**增加导出结果的功能**~~

**~~彩色输出？感觉挺好看的...就是有点花眼睛感觉🧐~~**

**还有啥功能想到了再加..**


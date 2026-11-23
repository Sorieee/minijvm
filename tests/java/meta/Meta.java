// Meta 是 M03 结构对拍用的固定测试类（常量池/属性覆盖面广）。
//
// 刻意覆盖的解析要点：
//   - 静态 final 常量（int / long / double / String）→ ConstantValue 属性
//   - long / double 字面量              → 常量池双槽位
//   - throws IOException 的方法          → Exceptions 属性
//   - 实现两个接口                       → 接口表
//   - 泛型 Comparable<Meta>              → Signature 属性（未知属性跳过的用例）
//   - 实例字段与 final 字段              → 字段表
//
// 编译：javac --release 8（保证 v52 字节码；字符串拼接走 StringBuilder，
// 不产生 invokedynamic）。期望值用 javap -v 人工记录后写进测试断言。

import java.io.IOException;

public class Meta implements Runnable, Comparable<Meta> {

    // ---- 静态 final 常量：编译期常量，各自带 ConstantValue 属性 ----

    public static final int INT_CONST = 42;
    public static final long LONG_CONST = 1234567890123L;
    public static final double DOUBLE_CONST = 3.14159;
    public static final String STRING_CONST = "hello";

    // ---- 实例字段：非 static，不会有 ConstantValue 属性 ----

    private int instanceField;
    private final String name;

    public Meta(String name) {
        this.name = name;
        this.instanceField = 1;
    }

    // ---- throws 子句：生成 Exceptions 属性 ----

    public void throwIOException() throws IOException {
        throw new IOException("boom");
    }

    // ---- 覆盖接口方法：接口表里有两个接口 ----

    @Override
    public void run() {
        this.instanceField++;
    }

    @Override
    public int compareTo(Meta other) {
        return this.name.compareTo(other.name);
    }

    // ---- 静态方法 + 实例方法：方法表的基本形态 ----

    public static int staticMethod(int a, int b) {
        return a + b;
    }

    public int instanceMethod() {
        return this.instanceField;
    }

    public static void main(String[] args) {
        System.out.println(INT_CONST + " " + LONG_CONST + " " + DOUBLE_CONST + " " + STRING_CONST);
    }
}

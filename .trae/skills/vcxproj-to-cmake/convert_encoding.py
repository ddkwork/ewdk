"""
递归扫描 SysTemp18 目录下的所有文本文件（.h/.c/.cpp/.asm/.inc/.rc/.inf/.txt 等），
将 GB2312/GBK 编码的文件转换为 UTF-8（带 BOM，兼容 MSVC 编译器）。
"""
import os
import sys

# 项目根目录
BASE_DIR = os.path.dirname(os.path.abspath(__file__))

# 需要处理的文件扩展名（文本文件）
TEXT_EXTENSIONS = {
    '.h', '.c', '.cpp', '.hpp', '.asm', '.inc', '.rc', '.inf',
    '.txt', '.py', '.sln', '.vcxproj', '.vcxproj.filters',
    '.gitattributes', '.gitignore',
}

# 需要跳过的目录
SKIP_DIRS = {'.git', '.svn', 'res'}

# 需要跳过的文件（二进制文件或无需转换的文件）
SKIP_FILES = {
    '.ico', '.png', '.jpg', '.jpeg', '.bmp', '.gif',
    '.exe', '.dll', '.sys', '.obj', '.lib', '.pdb',
}


def should_skip(filename: str) -> bool:
    """检查文件是否应跳过"""
    _, ext = os.path.splitext(filename)
    return ext.lower() in SKIP_FILES or ext.lower() not in TEXT_EXTENSIONS


def try_decode(data: bytes, encodings: list[str]) -> str | None:
    """尝试用多种编码解码"""
    for enc in encodings:
        try:
            return data.decode(enc)
        except (UnicodeDecodeError, LookupError):
            continue
    return None


def convert_file(filepath: str) -> bool:
    """将文件转换为 UTF-8 无 BOM。返回 True 表示已转换。"""
    if should_skip(filepath):
        return False

    try:
        with open(filepath, 'rb') as f:
            raw = f.read()
    except OSError as e:
        print(f"  [!] 读取失败: {filepath} -> {e}")
        return False

    # 获取文件扩展名
    _, ext = os.path.splitext(filepath)

    # 1) UTF-16LE (BOM: \xff\xfe) → 跳过（.rc 资源文件标准格式，VS 依赖此格式）
    if raw[:2] == b'\xff\xfe':
        return False

    # 2) 已经是 UTF-8 BOM → 移除 BOM
    if raw[:3] == b'\xef\xbb\xbf':
        with open(filepath, 'wb') as f:
            f.write(raw[3:])
        return True

    # 3) 纯 ASCII → 无需转换
    try:
        raw.decode('ascii')
        return False
    except UnicodeDecodeError:
        pass

    # 4) 已经是 UTF-8 但无 BOM → 无需转换
    try:
        raw.decode('utf-8')
        return False
    except UnicodeDecodeError:
        pass

    # 5) 尝试 GB2312/GBK → 转 UTF-8 无 BOM
    text = try_decode(raw, ['gbk', 'gb2312', 'gb18030'])
    if text is not None:
        with open(filepath, 'wb') as f:
            f.write(text.encode('utf-8'))
        return True

    return False


def main():
    converted_count = 0
    total_count = 0

    for root, dirs, files in os.walk(BASE_DIR):
        # 跳过不需要的目录
        dirs[:] = [d for d in dirs if d not in SKIP_DIRS and not d.startswith('.')]

        for filename in files:
            filepath = os.path.join(root, filename)
            relpath = os.path.relpath(filepath, BASE_DIR)

            if should_skip(filename):
                continue

            total_count += 1
            if convert_file(filepath):
                converted_count += 1
                print(f"  [OK] {relpath}")

    print(f"\n完成！共扫描 {total_count} 个文本文件，转换了 {converted_count} 个文件。")


if __name__ == '__main__':
    main()

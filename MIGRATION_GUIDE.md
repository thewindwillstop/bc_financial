# 🚀 快速迁移指南 - WSL2 → Linux

## 📦 需要迁移的内容

### ✅ 保留的核心文件

```
bc_financial/
├── backend/              # Go 后端代码（核心）
├── contracts/            # Solidity 智能合约
├── database/             # 数据库 Schema
├── fisco/                # FISCO BCOS 节点（已部署）
├── README.md             # 项目说明
├── TODO.md               # 待办清单
├── PROJECT_GUIDE.md      # 项目指南
└── idea.md               # 设计笔记
```

### ❌ 已删除的文件

- fabric-samples/      # Fabric 示例（可重新下载）
- fabric_chaincode/    # Fabric 代码（已不用）
- deploy_fabric.sh     # Fabric 部署脚本
- test_fabric.sh       # Fabric 测试脚本
- 其他 Fabric 相关文档

---

## 📤 迁移步骤

### 方法 1: Git 推送（推荐）

```bash
# 在 WSL2 中
cd /home/lin123456/colloge_project/bc_financial
git add .
git commit -m "准备迁移到 Linux 系统"
git push origin main
```

然后在 Linux 系统中：
```bash
git clone <your-repo-url> bc_financial
cd bc_financial
```

### 方法 2: 压缩包

```bash
# 在 WSL2 中
cd /home/lin123456/colloge_project
tar -czf bc_financial.tar.gz bc_financial/

# 然后通过 scp、U盘或其他方式传输到 Linux
```

在 Linux 中解压：
```bash
tar -xzf bc_financial.tar.gz
cd bc_financial
```

### 方法 3: rsync（如果有网络连接）

```bash
# 在 WSL2 中
rsync -avz --progress \
  /home/lin123456/colloge_project/bc_financial/ \
  <user>@<linux-ip>:/home/user/bc_financial/
```

---

## ⚠️ 注意事项

### 1. 不需要迁移的文件

- `fisco/nodes/` 中的日志文件（.log）
- `backend/bin/` 编译后的二进制文件
- Go 缓存文件
- 临时文件

### 2. 需要在 Linux 重新生成

- Go 编译的二进制文件
- 证书文件（如果需要）
- 配置文件的绝对路径

### 3. 配置文件需要修改

**backend/configs/config.yaml**:
```yaml
# 检查这些路径是否正确
database:
  mysql:
    host: localhost    # 可能需要改

blockchain:
  config_file: ./configs/fisco_config.toml  # 确保路径正确
```

---

## 🎯 到达 Linux 后的第一件事

1. **解压/克隆项目**
2. **查看 TODO.md**
3. **按照 TODO.md 中的"阶段 1"开始**

```bash
cat TODO.md
```

---

## 📞 如果遇到问题

- 查看 TODO.md 的详细步骤
- 查看 PROJECT_GUIDE.md 的项目说明
- 随时回来问我

---

**祝迁移顺利！** 🚀

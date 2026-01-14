/**
 * @title Reconciliation
 * @dev 金融交易数据分布式对账与溯源智能合约
 * @notice 基于"哈希上链,链上碰撞"机制实现隐私保护的对账系统
 */
pragma solidity ^0.6.10;

pragma experimental ABIEncoderV2;

/**
 * @dev 交易状态枚举
 */
enum TxStatus {
    PENDING,        // 待上链
    UPLOADED,       // 已上链(单方)
    MATCHED,        // 对账成功
    MISMATCH,       // 对账失败
    DISPUTED        // 存在争议
}

/**
 * @dev 交易记录结构体
 */
struct Transaction {
    bytes32 txHash;           // 交易数据哈希
    address uploader;         // 上传者地址(代表所属机构)
    uint256 timestamp;        // 上传时间戳
    TxStatus status;          // 交易状态
    address counterparty;     // 交易对手方地址
    uint256 matchHeight;      // 对账成功时的区块高度
}

/**
 * @dev 对账事件
 * @param bizId 业务流水号
 * @param status 对账状态
 * @param uploader 交易上传方
 * @param counterparty 交易对手方
 * @param blockHeight 当前区块高度
 */
event ReconciliationEvent(
    bytes32 indexed bizId,
    TxStatus status,
    address indexed uploader,
    address indexed counterparty,
    uint256 blockHeight
);

/**
 * @dev 数据上链事件
 * @param bizId 业务流水号
 * @param dataHash 数据哈希
 * @param uploader 上传者
 */
event DataUploaded(
    bytes32 indexed bizId,
    bytes32 dataHash,
    address indexed uploader,
    uint256 timestamp
);

/**
 * @dev 机构信息
 */
struct Institution {
    string name;              // 机构名称
    address addr;             // 机构地址
    bool isRegistered;        // 是否已注册
    uint256 uploadCount;      // 上传交易数量
    uint256 matchedCount;     // 对账成功数量
}

/**
 * @title Reconciliation
 * @author 毕业设计项目
 */
contract Reconciliation {

    // ========== 状态变量 ==========

    address public owner;                                     // 合约所有者(系统管理员)
    bool public paused;                                      // 合约暂停状态
    uint256 public txCount;                                  // 总交易数
    uint256 public matchedCount;                             // 总对账成功数

    // 业务流水号 => 交易记录映射
    mapping(bytes32 => Transaction) public transactions;

    // 业务流水号 => 是否存在映射
    mapping(bytes32 => bool) public txExists;

    // 机构地址 => 机构信息映射
    mapping(address => Institution) public institutions;

    // 已注册的机构地址列表
    address[] public institutionList;

    // ========== 修饰符 ==========

    /**
     * @dev 仅合约所有者可调用
     */
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    /**
     * @dev 仅已注册机构可调用
     */
    modifier onlyRegistered() {
        require(
            institutions[msg.sender].isRegistered,
            "Institution not registered"
        );
        _;
    }

    // ========== 构造函数 ==========

    /**
     * @dev 构造函数
     */
    constructor() public {
        owner = msg.sender;
        paused = false;
        txCount = 0;
        matchedCount = 0;
    }

    // ========== 核心功能函数 ==========

    /**
     * @dev 注册金融机构
     * @param name 机构名称
     * @param addr 机构地址
     */
    function registerInstitution(string memory name, address addr)
        public
        onlyOwner
    {
        require(!institutions[addr].isRegistered, "Institution already registered");

        institutions[addr] = Institution({
            name: name,
            addr: addr,
            isRegistered: true,
            uploadCount: 0,
            matchedCount: 0
        });

        institutionList.push(addr);

        emit InstitutionRegistered(addr, name, block.timestamp);
    }

    /**
     * @dev 批量注册机构
     * @param names 机构名称数组
     * @param addrs 机构地址数组
     */
    function batchRegisterInstitutions(
        string[] memory names,
        address[] memory addrs
    ) public onlyOwner {
        require(names.length == addrs.length, "Arrays length mismatch");

        for (uint256 i = 0; i < addrs.length; i++) {
            if (!institutions[addrs[i]].isRegistered) {
                institutions[addrs[i]] = Institution({
                    name: names[i],
                    addr: addrs[i],
                    isRegistered: true,
                    uploadCount: 0,
                    matchedCount: 0
                });
                institutionList.push(addrs[i]);
                emit InstitutionRegistered(addrs[i], names[i], now);
            }
        }
    }

    /**
     * @dev 上传交易数据哈希上链
     * @param bizId 业务流水号(作为key)
     * @param dataHash 数据哈希值 SHA256(流水号+金额+盐)
     */
    function uploadTransaction(bytes32 bizId, bytes32 dataHash)
        public
        onlyRegistered
        whenNotPaused
        returns (bool)
    {
        // 检查是否已存在
        if (!txExists[bizId]) {
            // 首次上传,创建记录
            transactions[bizId] = Transaction({
                txHash: dataHash,
                uploader: msg.sender,
                timestamp: now,
                status: TxStatus.UPLOADED,
                counterparty: address(0),
                matchHeight: 0
            });

            txExists[bizId] = true;
            txCount++;
            institutions[msg.sender].uploadCount++;

            emit DataUploaded(bizId, dataHash, msg.sender, block.timestamp);

            return true;
        } else {
            // 交易已存在,执行对账逻辑
            Transaction storage existingTx = transactions[bizId];

            // 检查是否是同一机构重复上传
            require(
                existingTx.uploader != msg.sender,
                "Transaction already uploaded by this institution"
            );

            // 执行哈希碰撞对账
            if (existingTx.txHash == dataHash) {
                // 哈希相同,对账成功
                existingTx.status = TxStatus.MATCHED;
                existingTx.counterparty = msg.sender;
                existingTx.matchHeight = block.number;

                // 更新机构统计
                institutions[existingTx.uploader].matchedCount++;
                institutions[msg.sender].matchedCount++;
                matchedCount++;

                emit ReconciliationEvent(
                    bizId,
                    TxStatus.MATCHED,
                    existingTx.uploader,
                    msg.sender,
                    block.number
                );

                return true;
            } else {
                // 哈希不同,对账失败(数据被篡改)
                existingTx.status = TxStatus.MISMATCH;
                existingTx.counterparty = msg.sender;

                emit ReconciliationEvent(
                    bizId,
                    TxStatus.MISMATCH,
                    existingTx.uploader,
                    msg.sender,
                    block.number
                );

                return false;
            }
        }
    }

    /**
     * @dev 批量上传交易
     * @param bizIds 业务流水号数组
     * @param dataHashes 数据哈希数组
     */
    function batchUploadTransactions(
        bytes32[] memory bizIds,
        bytes32[] memory dataHashes
    ) public onlyRegistered whenNotPaused returns (uint256 successCount) {
        require(
            bizIds.length == dataHashes.length,
            "Arrays length mismatch"
        );

        successCount = 0;

        for (uint256 i = 0; i < bizIds.length; i++) {
            if (uploadTransaction(bizIds[i], dataHashes[i])) {
                successCount++;
            }
        }
    }

    // ========== 查询函数 ==========

    /**
     * @dev 查询交易详情
     * @param bizId 业务流水号
     */
    function getTransaction(bytes32 bizId)
        public
        view
        returns (
            bytes32 dataHash,
            address uploader,
            uint256 timestamp,
            TxStatus status,
            address counterparty,
            uint256 matchHeight
        )
    {
        require(txExists[bizId], "Transaction does not exist");

        Transaction memory tx = transactions[bizId];
        return (
            tx.txHash,
            tx.uploader,
            tx.timestamp,
            tx.status,
            tx.counterparty,
            tx.matchHeight
        );
    }

    /**
     * @dev 查询机构信息
     * @param addr 机构地址
     */
    function getInstitution(address addr)
        public
        view
        returns (
            string memory name,
            address institutionAddr,
            bool isRegistered,
            uint256 uploadCount,
            uint256 matchedCount
        )
    {
        Institution memory inst = institutions[addr];
        return (
            inst.name,
            inst.addr,
            inst.isRegistered,
            inst.uploadCount,
            inst.matchedCount
        );
    }

    /**
     * @dev 获取所有已注册机构数量
     */
    function getInstitutionCount() public view returns (uint256) {
        return institutionList.length;
    }

    /**
     * @dev 查询对账统计信息
     */
    function getStatistics()
        public
        view
        returns (
            uint256 totalTx,
            uint256 totalMatched,
            uint256 matchRate, // 匹配率(基数为10000)
            uint256 institutionCount
        )
    {
        uint256 rate = 0;
        if (txCount > 0) {
            rate = (matchedCount * 10000) / txCount;
        }

        return (txCount, matchedCount, rate, institutionList.length);
    }

    /**
     * @dev 验证交易哈希是否存在(用于前端点击验证)
     * @param bizId 业务流水号
     * @param dataHash 待验证的哈希值
     */
    function verifyTransaction(bytes32 bizId, bytes32 dataHash)
        public
        view
        returns (bool isValid, TxStatus status)
    {
        if (!txExists[bizId]) {
            return (false, TxStatus.PENDING);
        }

        Transaction memory tx = transactions[bizId];
        return (tx.txHash == dataHash, tx.status);
    }

    // ========== 管理员函数 ==========

    /**
     * @dev 转移合约所有权
     */
    function transferOwnership(address newOwner) public onlyOwner {
        require(newOwner != address(0), "Invalid address");
        owner = newOwner;
    }

    /**
     * @dev 暂停合约(紧急情况下使用)
     */
    function pause() public onlyOwner {
        paused = true;
        emit ContractPaused(msg.sender, true);
    }

    function unpause() public onlyOwner {
        paused = false;
        emit ContractPaused(msg.sender, false);
    }

    // ========== 事件定义 ==========

    event InstitutionRegistered(
        address indexed institutionAddr,
        string name,
        uint256 timestamp
    );

    event OwnershipTransferred(
        address indexed previousOwner,
        address indexed newOwner
    );

    event ContractPaused(address indexed admin, bool paused);
}

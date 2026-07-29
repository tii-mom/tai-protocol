// TAI Protocol - TAI Jetton master and wallet
// Standard: TEP-74 (Jettons)
// Language: Tolk 1.0
//
// This file contains both roles required by TEP-74. The same compiled code
// is deployed once as the master (role=0) and as deterministic wallet items
// (role=1). The master never stores holder balances; every wallet stores its
// own balance and processes transfer/burn messages.
//
// Standard inbound message formats:
// 1. transfer#0f8a7ea5 query_id:uint64 amount:coins destination:address
//    response_destination:address? custom_payload:(Maybe ^Cell)
//    forward_amount:coins forward_payload:(Either Cell ^Cell).
// 2. internal_transfer#178d4519 query_id:uint64 amount:coins
//    from:address? response_destination:address? forward_amount:coins
//    forward_payload:(Either Cell ^Cell).
// 3. burn#595f07bc query_id:uint64 amount:coins
//    response_destination:address? custom_payload:(Maybe ^Cell).
// 4. burn_notification#7bdd97de query_id:uint64 amount:coins
//    sender:address response_destination:address?.
// 5. discover#2c76b973 query_id:uint64 ownerAddress:address
//    includeOwnerAddress:bool.
// 6. mint#00000015 queryId:uint64 recipient:address gramAmount:coins
//    internalTransfer:^Cell<InternalTransfer> (owner only).
// 7. breeding_burn#00000017 queryId:uint64 feeAmount:coins
//    burnAmount:coins responseDestination:address? (breeding only).
//
// Owner/admin messages are change_owner#00000003, change_content#00000004,
// and set_breeding#00000018. TAI uses 9 decimals. The supply cap is 1B TAI.
// Protocol fee burning is expressed as exactly FEE_BURN_PERCENT of the fee;
// breeding uses the explicit breeding_burn path so the master can authorize
// the breeding contract without granting it owner/mint privileges.

const ROLE_MASTER = 0
const ROLE_WALLET = 1

const TOTAL_SUPPLY_CAP = 1_000_000_000_000_000_000
const FEE_BURN_PERCENT = 50
const MAX_FORWARD_PAYLOAD_BITS = 1023

const OP_TRANSFER = 0x0f8a7ea5
const OP_INTERNAL_TRANSFER = 0x178d4519
const OP_TRANSFER_NOTIFICATION = 0x7362d09c
const OP_EXCESSES = 0xd53276db
const OP_BURN = 0x595f07bc
const OP_BURN_NOTIFICATION = 0x7bdd97de
const OP_DISCOVER_WALLET = 0x2c76b973
const OP_WALLET_ADDRESS = 0xd1735400
const OP_MINT = 0x00000015
const OP_CHANGE_OWNER = 0x00000003
const OP_CHANGE_CONTENT = 0x00000004
const OP_BREEDING_BURN = 0x00000017
const OP_SET_BREEDING = 0x00000018

const ERR_INVALID_OPERATION = 700
const ERR_NOT_FROM_OWNER = 701
const ERR_NOT_FROM_MINTER = 702
const ERR_NOT_FROM_ADMIN = 703
const ERR_NOT_FROM_BREEDING = 704
const ERR_UNAUTHORIZED_BURN = 705
const ERR_NOT_ENOUGH_BALANCE = 706
const ERR_SUPPLY_CAP = 707
const ERR_NOT_ENOUGH_TON = 708
const ERR_INVALID_BURN_AMOUNT = 709
const ERR_INVALID_WALLET = 710
const ERR_WRONG_WORKCHAIN = 711
const MIN_WALLET_MESSAGE_VALUE = 20_000_000 // 0.02 TON in nanotons

type ForwardPayload = RemainingBitsAndRefs

struct TaiMasterStorage {
    totalSupply: coins
    ownerAddress: address
    breedingAddress: address?
    content: cell
    walletCode: cell
}

struct TaiWalletStorage {
    ownerAddress: address
    minterAddress: address
    jettonBalance: coins
    content: cell
}

struct TaiJettonStorage {
    role: uint8
    masterData: Cell<TaiMasterStorage>?
    walletData: Cell<TaiWalletStorage>?
}

struct (0x0f8a7ea5) TransferJettons {
    queryId: uint64
    jettonAmount: coins
    destination: address
    responseDestination: address?
    customPayload: cell?
    forwardAmount: coins
    forwardPayload: ForwardPayload
}

struct (0x178d4519) InternalTransferJettons {
    queryId: uint64
    jettonAmount: coins
    transferInitiator: address?
    sendExcessesTo: address?
    forwardAmount: coins
    forwardPayload: ForwardPayload
}

struct (0x7362d09c) TransferNotification {
    queryId: uint64
    jettonAmount: coins
    transferInitiator: address?
    forwardPayload: ForwardPayload
}

struct (0xd53276db) Excesses {
    queryId: uint64
}

struct (0x595f07bc) BurnJettons {
    queryId: uint64
    jettonAmount: coins
    responseDestination: address?
    customPayload: cell?
}

struct (0x7bdd97de) BurnNotification {
    queryId: uint64
    jettonAmount: coins
    burnInitiator: address
    responseDestination: address?
}

struct (0x2c76b973) DiscoverWallet {
    queryId: uint64
    ownerAddress: address
    includeOwnerAddress: bool
}

struct (0xd1735400) WalletAddressReply {
    queryId: uint64
    walletAddress: address?
    ownerAddress: Cell<address>?
}

struct (0x00000015) MintJettons {
    queryId: uint64
    mintRecipient: address
    gramAmount: coins
    internalTransfer: Cell<InternalTransferJettons>
}

struct (0x00000003) ChangeOwner {
    queryId: uint64
    newOwnerAddress: address
}

struct (0x00000004) ChangeContent {
    queryId: uint64
    newContent: cell
}

struct (0x00000017) BreedingBurn {
    queryId: uint64
    feeAmount: coins
    burnAmount: coins
    responseDestination: address?
}

struct (0x00000018) SetBreeding {
    queryId: uint64
    breedingAddress: address
}

type TaiIncomingMessage =
    | TransferJettons
    | InternalTransferJettons
    | TransferNotification
    | Excesses
    | BurnJettons
    | BurnNotification
    | DiscoverWallet
    | MintJettons
    | ChangeOwner
    | ChangeContent
    | BreedingBurn
    | SetBreeding

type TaiBounceMessage = InternalTransferJettons | BurnNotification

contract TaiJetton {
    storage: TaiJettonStorage
    incomingMessages: TaiIncomingMessage
}

fun TaiJettonStorage.load() {
    return TaiJettonStorage.fromCell(contract.getData());
}

fun TaiJettonStorage.save(self) {
    contract.setData(self.toCell());
}

fun emptyCell(): cell {
    return beginCell().endCell();
}

fun makeWalletState(
    ownerAddress: address,
    minterAddress: address,
    content: cell,
): TaiJettonStorage {
    val walletData: TaiWalletStorage = {
        ownerAddress,
        minterAddress,
        jettonBalance: 0,
        content,
    };
    return {
        role: ROLE_WALLET,
        masterData: null,
        walletData: walletData.toCell(),
    };
}

fun calcDeployedWallet(
    ownerAddress: address,
    minterAddress: address,
    content: cell,
    walletCode: cell,
): AutoDeployAddress {
    val walletState = makeWalletState(
        ownerAddress,
        minterAddress,
        content,
    );
    return { stateInit: { code: walletCode, data: walletState.toCell() } };
}

fun calcWalletAddress(
    ownerAddress: address,
    minterAddress: address,
    content: cell,
    walletCode: cell,
): address {
    return calcDeployedWallet(
        ownerAddress,
        minterAddress,
        content,
        walletCode,
    ).calculateAddress();
}

fun protocolBurnAmount(feeAmount: coins): coins {
    return feeAmount * FEE_BURN_PERCENT / 100;
}

fun sendWalletExcesses(destination: address, queryId: uint64) {
    val message = createMessage({
        bounce: BounceMode.NoBounce,
        dest: destination,
        value: 0,
        body: Excesses { queryId },
    });
    message.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
}

fun handleMasterMessage(
    senderAddress: address,
    msg: TaiIncomingMessage,
    mutate storage: TaiMasterStorage,
) {
    match (msg) {
        MintJettons => {
            assert (senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_ADMIN;
            val transfer = msg.internalTransfer.load();
            assert (transfer.transferInitiator == null) throw ERR_INVALID_OPERATION;
            assert (transfer.jettonAmount > 0) throw ERR_INVALID_OPERATION;
            assert (
                storage.totalSupply + transfer.jettonAmount <= TOTAL_SUPPLY_CAP
            ) throw ERR_SUPPLY_CAP;

            storage.totalSupply += transfer.jettonAmount;
            val deploy = createMessage({
                bounce: BounceMode.Only256BitsOfBody,
                dest: calcDeployedWallet(
                    msg.mintRecipient,
                    contract.getAddress(),
                    storage.content,
                    storage.walletCode,
                ),
                value: msg.gramAmount,
                body: msg.internalTransfer,
            });
            deploy.send(SEND_MODE_PAY_FEES_SEPARATELY);
        }

        BurnNotification => {
            assert (
                senderAddress ==
                calcWalletAddress(
                    msg.burnInitiator,
                    contract.getAddress(),
                    storage.content,
                    storage.walletCode,
                )
            ) throw ERR_UNAUTHORIZED_BURN;
            assert (storage.totalSupply >= msg.jettonAmount) throw ERR_INVALID_BURN_AMOUNT;

            storage.totalSupply -= msg.jettonAmount;
            if (msg.responseDestination != null) {
                sendWalletExcesses(msg.responseDestination!, msg.queryId);
            }
        }

        DiscoverWallet => {
            assert (msg.ownerAddress.getWorkchain() == BASECHAIN) throw ERR_WRONG_WORKCHAIN;
            val respondOwner = msg.includeOwnerAddress
                ? msg.ownerAddress.toCell()
                : null as Cell<address>?;
            val walletAddress = calcWalletAddress(
                msg.ownerAddress,
                contract.getAddress(),
                storage.content,
                storage.walletCode,
            );
            val response = createMessage({
                bounce: BounceMode.Only256BitsOfBody,
                dest: senderAddress,
                value: 0,
                body: WalletAddressReply {
                    queryId: msg.queryId,
                    walletAddress,
                    ownerAddress: respondOwner,
                },
            });
            response.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }

        ChangeOwner => {
            assert (senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_ADMIN;
            storage.ownerAddress = msg.newOwnerAddress;
        }

        ChangeContent => {
            assert (senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_ADMIN;
            storage.content = msg.newContent;
        }

        SetBreeding => {
            assert (senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_ADMIN;
            storage.breedingAddress = msg.breedingAddress;
        }

        BreedingBurn => {
            assert (storage.breedingAddress != null) throw ERR_NOT_FROM_BREEDING;
            assert (senderAddress == storage.breedingAddress!) throw ERR_NOT_FROM_BREEDING;
            assert (
                msg.burnAmount == protocolBurnAmount(msg.feeAmount)
            ) throw ERR_INVALID_BURN_AMOUNT;
            assert (msg.burnAmount > 0) throw ERR_INVALID_BURN_AMOUNT;

            val breedingWallet = calcDeployedWallet(
                storage.breedingAddress!,
                contract.getAddress(),
                storage.content,
                storage.walletCode,
            );
            val burn = createMessage({
                bounce: BounceMode.Only256BitsOfBody,
                dest: breedingWallet,
                value: 0,
                body: BurnJettons {
                    queryId: msg.queryId,
                    jettonAmount: msg.burnAmount,
                    responseDestination: msg.responseDestination,
                    customPayload: null,
                },
            });
            burn.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }

        TransferJettons => {
            throw ERR_INVALID_OPERATION;
        }
        InternalTransferJettons => {
            throw ERR_INVALID_OPERATION;
        }
        TransferNotification => {
            throw ERR_INVALID_OPERATION;
        }
        Excesses => {
            throw ERR_INVALID_OPERATION;
        }
        BurnJettons => {
            throw ERR_INVALID_OPERATION;
        }
    }
}

fun handleWalletMessage(
    senderAddress: address,
    valueCoins: coins,
    originalForwardFee: coins,
    msg: TaiIncomingMessage,
    mutate storage: TaiWalletStorage,
) {
    match (msg) {
        TransferJettons => {
            assert (senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_OWNER;
            assert (msg.destination.getWorkchain() == BASECHAIN) throw ERR_WRONG_WORKCHAIN;
            assert (msg.jettonAmount > 0) throw ERR_INVALID_OPERATION;
            assert (storage.jettonBalance >= msg.jettonAmount) throw ERR_NOT_ENOUGH_BALANCE;
            assert (
                valueCoins >=
                msg.forwardAmount + (2 * originalForwardFee) + MIN_WALLET_MESSAGE_VALUE
            ) throw ERR_NOT_ENOUGH_TON;

            storage.jettonBalance -= msg.jettonAmount;
            val destinationWallet = calcDeployedWallet(
                msg.destination,
                storage.minterAddress,
                storage.content,
                contract.getCode(),
            );
            val transfer = createMessage({
                bounce: BounceMode.Only256BitsOfBody,
                dest: destinationWallet,
                value: 0,
                body: InternalTransferJettons {
                    queryId: msg.queryId,
                    jettonAmount: msg.jettonAmount,
                    transferInitiator: storage.ownerAddress,
                    sendExcessesTo: msg.responseDestination,
                    forwardAmount: msg.forwardAmount,
                    forwardPayload: msg.forwardPayload,
                },
            });
            transfer.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }

        InternalTransferJettons => {
            if (msg.transferInitiator == null) {
                assert (senderAddress == storage.minterAddress) throw ERR_NOT_FROM_MINTER;
            } else {
                assert (
                    senderAddress ==
                    calcWalletAddress(
                        msg.transferInitiator!,
                        storage.minterAddress,
                        storage.content,
                        contract.getCode(),
                    )
                ) throw ERR_INVALID_WALLET;
            }
            assert (msg.jettonAmount > 0) throw ERR_INVALID_OPERATION;

            storage.jettonBalance += msg.jettonAmount;
            if (msg.forwardAmount != 0) {
                val notification = createMessage({
                    bounce: BounceMode.NoBounce,
                    dest: storage.ownerAddress,
                    value: msg.forwardAmount,
                    body: TransferNotification {
                        queryId: msg.queryId,
                        jettonAmount: msg.jettonAmount,
                        transferInitiator: msg.transferInitiator,
                        forwardPayload: msg.forwardPayload,
                    },
                });
                notification.send(SEND_MODE_PAY_FEES_SEPARATELY);
            }

            if (msg.sendExcessesTo != null) {
                sendWalletExcesses(msg.sendExcessesTo!, msg.queryId);
            }
        }

        BurnJettons => {
            assert (senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_OWNER;
            assert (msg.jettonAmount > 0) throw ERR_INVALID_OPERATION;
            assert (storage.jettonBalance >= msg.jettonAmount) throw ERR_NOT_ENOUGH_BALANCE;

            storage.jettonBalance -= msg.jettonAmount;
            val notification = createMessage({
                bounce: BounceMode.Only256BitsOfBody,
                dest: storage.minterAddress,
                value: 0,
                body: BurnNotification {
                    queryId: msg.queryId,
                    jettonAmount: msg.jettonAmount,
                    burnInitiator: storage.ownerAddress,
                    responseDestination: msg.responseDestination,
                },
            });
            notification.send(
                SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE | SEND_MODE_BOUNCE_ON_ACTION_FAIL,
            );
        }

        TransferNotification => {
            throw ERR_INVALID_OPERATION;
        }
        Excesses => {
            throw ERR_INVALID_OPERATION;
        }
        BurnNotification => {
            throw ERR_INVALID_OPERATION;
        }
        DiscoverWallet => {
            throw ERR_INVALID_OPERATION;
        }
        MintJettons => {
            throw ERR_INVALID_OPERATION;
        }
        ChangeOwner => {
            throw ERR_INVALID_OPERATION;
        }
        ChangeContent => {
            throw ERR_INVALID_OPERATION;
        }
        BreedingBurn => {
            throw ERR_INVALID_OPERATION;
        }
        SetBreeding => {
            throw ERR_INVALID_OPERATION;
        }
    }
}

fun onInternalMessage(in: InMessage) {
    if (in.body.isEmpty()) {
        return;
    }
    val msg = TaiIncomingMessage.fromSlice(in.body);
    var storage = lazy TaiJettonStorage.load();

    if (storage.role == ROLE_MASTER) {
        assert (storage.masterData != null) throw ERR_INVALID_OPERATION;
        var master = storage.masterData!.load();
        handleMasterMessage(
            in.senderAddress,
            msg,
            mutate master,
        );
        storage.masterData = master.toCell();
        storage.save();
    } else if (storage.role == ROLE_WALLET) {
        assert (storage.walletData != null) throw ERR_INVALID_OPERATION;
        var wallet = storage.walletData!.load();
        handleWalletMessage(
            in.senderAddress,
            in.valueCoins,
            in.originalForwardFee,
            msg,
            mutate wallet,
        );
        storage.walletData = wallet.toCell();
        storage.save();
    } else {
        throw ERR_INVALID_OPERATION;
    }
}

fun onBouncedMessage(in: InMessageBounced) {
    var storage = lazy TaiJettonStorage.load();
    if (storage.role != ROLE_WALLET) {
        return;
    }
    assert (storage.walletData != null) throw ERR_INVALID_OPERATION;
    var wallet = storage.walletData!.load();

    in.bouncedBody.skipBouncedPrefix();
    val msg = lazy TaiBounceMessage.fromSlice(in.bouncedBody);
    match (msg) {
        InternalTransferJettons => {
            wallet.jettonBalance += msg.jettonAmount;
        }
        BurnNotification => {
            wallet.jettonBalance += msg.jettonAmount;
        }
        else => {
            return;
        }
    }
    storage.walletData = wallet.toCell();
    storage.save();
}

struct JettonDataReply {
    totalSupply: coins
    mintable: bool
    adminAddress: address
    jettonContent: cell
    jettonWalletCode: cell
}

get fun get_jetton_data(): JettonDataReply {
    val storage = lazy TaiJettonStorage.load();
    assert (storage.role == ROLE_MASTER) throw ERR_INVALID_OPERATION;
    assert (storage.masterData != null) throw ERR_INVALID_OPERATION;
    val master = storage.masterData!.load();
    return {
        totalSupply: master.totalSupply,
        mintable: master.totalSupply < TOTAL_SUPPLY_CAP,
        adminAddress: master.ownerAddress,
        jettonContent: master.content,
        jettonWalletCode: master.walletCode,
    };
}

get fun get_wallet_address(ownerAddress: address): address {
    val storage = lazy TaiJettonStorage.load();
    assert (storage.role == ROLE_MASTER) throw ERR_INVALID_OPERATION;
    assert (storage.masterData != null) throw ERR_INVALID_OPERATION;
    val master = storage.masterData!.load();
    return calcWalletAddress(
        ownerAddress,
        contract.getAddress(),
        master.content,
        master.walletCode,
    );
}

struct JettonWalletDataReply {
    jettonBalance: coins
    ownerAddress: address
    minterAddress: address
    jettonWalletCode: cell
}

get fun get_wallet_data(): JettonWalletDataReply {
    val storage = lazy TaiJettonStorage.load();
    assert (storage.role == ROLE_WALLET && storage.walletData != null) throw ERR_INVALID_OPERATION;
    val wallet = storage.walletData!.load();
    return {
        jettonBalance: wallet.jettonBalance,
        ownerAddress: wallet.ownerAddress,
        minterAddress: wallet.minterAddress,
        jettonWalletCode: contract.getCode(),
    };
}

get fun get_fee_burn_amount(feeAmount: coins): coins {
    return protocolBurnAmount(feeAmount);
}

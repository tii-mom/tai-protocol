// TAI Protocol - Pet NFT item contract
// Standard: TEP-62 NFT item / TEP-64 metadata
// Language: Tolk 1.0
//
// Incoming message formats:
// 1. transfer#5fcc3d14 query_id:uint64 new_owner:address
//    response_destination:address custom_payload:(Maybe ^Cell)
//    forward_amount:coins forward_payload:(Either Cell ^Cell).
// 2. get_static_data#2fcb26a2 query_id:uint64.
// 3. mint#00000001 ownerAddress:address metadataUri:^Cell
//    petData:^Cell<PetData> breedingContract:address. This initializes a
//    newly deployed item and may only be sent by collection/factory.
// 4. get_pet_data#00000005 queryId:uint64.
// 5. set_breed_cooldown#00000004 queryId:uint64 cooldownUntil:uint64.
// 6. update_pet_data#00000003 petData:^Cell<PetData>.
//
// `metadataUri` is a TEP-64 off-chain content cell. Its root must contain
// 0x01 followed by the URI of the platform API; JSON attributes stay off-chain.
// Outgoing standard messages are ownership_assigned#05138d91,
// excesses#d53276db, and report_static_data#8b771735.

const OP_MINT = 0x00000001
const OP_UPDATE_PET_DATA = 0x00000003
const OP_SET_BREED_COOLDOWN = 0x00000004
const OP_GET_PET_DATA = 0x00000005

const OP_TRANSFER = 0x5fcc3d14
const OP_GET_STATIC_DATA = 0x2fcb26a2
const OP_OWNERSHIP_ASSIGNED = 0x05138d91
const OP_EXCESSES = 0xd53276db
const OP_REPORT_STATIC_DATA = 0x8b771735

const MIN_NFT_STORAGE = 50_000_000 // 0.05 TON in nanotons

const ERR_NOT_FROM_COLLECTION = 401
const ERR_NOT_FROM_OWNER = 402
const ERR_NOT_FROM_BREEDING = 403
const ERR_NOT_INITIALIZED = 404
const ERR_BREED_COOLDOWN = 405
const ERR_NOT_ENOUGH_TON = 406
const ERR_INVALID_PET_DATA = 407
const ERR_INVALID_MESSAGE = 408

// The URI is a ref cell, because a Tolk `slice` cannot be auto-serialized.
// The same URI is returned as the NFT individual content in get_nft_data().
struct PetData {
    ownerAddress: address
    metadataUri: cell
    species: uint8
    quality: uint8
    generation: uint16
    growthRate: uint16
    aptAtk: uint8
    aptDef: uint8
    aptSpd: uint8
    aptInt: uint8
    skillSlots: uint8
    personality: uint8
    level: uint16
    exp: coins
    breedCooldownUntil: uint64
    totalEarned: coins
}

struct (0x00000001) NftMint {
    ownerAddress: address
    metadataUri: cell
    petData: Cell<PetData>
    breedingContract: address
}

struct (0x5fcc3d14) NftTransfer {
    queryId: uint64
    newOwnerAddress: address
    responseDestination: address
    customPayload: cell?
    forwardAmount: coins
    forwardPayload: RemainingBitsAndRefs
}

struct (0x2fcb26a2) NftGetStaticData {
    queryId: uint64
}

struct (0x00000005) GetPetDataRequest {
    queryId: uint64
}

struct (0x00000004) SetBreedCooldown {
    queryId: uint64
    cooldownUntil: uint64
}

struct (0x00000003) UpdatePetData {
    petData: Cell<PetData>
}

struct (0x05138d91) NftOwnershipAssigned {
    queryId: uint64
    previousOwnerAddress: address
    forwardPayload: RemainingBitsAndRefs
}

struct (0xd53276db) NftExcesses {
    queryId: uint64
}

struct (0x8b771735) NftReportStaticData {
    queryId: uint64
    itemIndex: uint256
    collectionAddress: address
}

struct (0x00000006) PetDataResponse {
    queryId: uint64
    petData: Cell<PetData>
}

struct PetNftStorageAtDeployment {
    itemIndex: uint64
    collectionAddress: address
}

struct PetNftStorage {
    itemIndex: uint64
    collectionAddress: address
    ownerAddress: address
    petData: Cell<PetData>
    metadataUri: cell
    breedingContract: address
    breedCooldownUntil: uint64
}

struct PetNftStorageLoader {
    contractData: slice
}

type PetNftIncomingMessage =
    | NftMint
    | NftTransfer
    | NftGetStaticData
    | GetPetDataRequest
    | SetBreedCooldown
    | UpdatePetData

contract PetNft {
    storage: PetNftStorage
    storageAtDeployment: PetNftStorageAtDeployment
    incomingMessages: PetNftIncomingMessage
}

fun PetNftStorageLoader.isInitialized(self) {
    // The initialized shape contains petData and metadataUri references;
    // deployment storage contains no references.
    return self.contractData.remainingRefsCount() >= 2;
}

fun PetNftStorageLoader.parseNotInitialized(self) {
    return PetNftStorageAtDeployment.fromSlice(self.contractData);
}

fun PetNftStorageLoader.parseInitialized(self) {
    return PetNftStorage.fromSlice(self.contractData);
}

fun startLoadingPetNftStorage(): PetNftStorageLoader {
    return { contractData: contract.getData().beginParse() };
}

fun PetNftStorage.save(self) {
    contract.setData(self.toCell());
}

fun emptyCell(): cell {
    return beginCell().endCell();
}

fun sendExcesses(destination: address, queryId: uint64) {
    val message = createMessage({
        bounce: BounceMode.NoBounce,
        dest: destination,
        value: 0,
        body: NftExcesses { queryId },
    });
    message.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
}

fun onInternalMessage(in: InMessage) {
    var loadingStorage = startLoadingPetNftStorage();

    if (!loadingStorage.isInitialized()) {
        val deployment = loadingStorage.parseNotInitialized();
        assert (in.senderAddress == deployment.collectionAddress) throw ERR_NOT_FROM_COLLECTION;

        val init = lazy NftMint.fromSlice(in.body);
        val pet = init.petData.load();
        assert (pet.ownerAddress == init.ownerAddress) throw ERR_INVALID_PET_DATA;

        val storage: PetNftStorage = {
            itemIndex: deployment.itemIndex,
            collectionAddress: deployment.collectionAddress,
            ownerAddress: init.ownerAddress,
            petData: init.petData,
            metadataUri: init.metadataUri,
            breedingContract: init.breedingContract,
            breedCooldownUntil: 0,
        };
        storage.save();
        return;
    }

    var storage = loadingStorage.parseInitialized();
    val msg = lazy PetNftIncomingMessage.fromSlice(in.body);

    match (msg) {
        NftTransfer => {
            assert (in.senderAddress == storage.ownerAddress) throw ERR_NOT_FROM_OWNER;
            assert (blockchain.now() >= storage.breedCooldownUntil) throw ERR_BREED_COOLDOWN;
            assert (
                in.valueCoins >= msg.forwardAmount + (2 * in.originalForwardFee) + MIN_NFT_STORAGE
            ) throw ERR_NOT_ENOUGH_TON;

            val previousOwner = storage.ownerAddress;
            storage.ownerAddress = msg.newOwnerAddress;

            var pet = storage.petData.load();
            pet.ownerAddress = msg.newOwnerAddress;
            storage.petData = pet.toCell();
            storage.save();

            if (msg.forwardAmount != 0) {
                val ownership = createMessage({
                    bounce: BounceMode.NoBounce,
                    dest: msg.newOwnerAddress,
                    value: msg.forwardAmount,
                    body: NftOwnershipAssigned {
                        queryId: msg.queryId,
                        previousOwnerAddress: previousOwner,
                        forwardPayload: msg.forwardPayload,
                    },
                });
                ownership.send(SEND_MODE_PAY_FEES_SEPARATELY);
            }

            sendExcesses(msg.responseDestination, msg.queryId);
        }

        NftGetStaticData => {
            val report = createMessage({
                bounce: BounceMode.NoBounce,
                dest: in.senderAddress,
                value: 0,
                body: NftReportStaticData {
                    queryId: msg.queryId,
                    itemIndex: storage.itemIndex as uint256,
                    collectionAddress: storage.collectionAddress,
                },
            });
            report.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }

        GetPetDataRequest => {
            assert (in.senderAddress == storage.breedingContract) throw ERR_NOT_FROM_BREEDING;

            val response = createMessage({
                bounce: BounceMode.NoBounce,
                dest: storage.breedingContract,
                value: 0,
                body: PetDataResponse {
                    queryId: msg.queryId,
                    petData: storage.petData,
                },
            });
            response.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
        }

        SetBreedCooldown => {
            assert (in.senderAddress == storage.breedingContract) throw ERR_NOT_FROM_BREEDING;

            storage.breedCooldownUntil = msg.cooldownUntil;
            var pet = storage.petData.load();
            pet.breedCooldownUntil = msg.cooldownUntil;
            storage.petData = pet.toCell();
            storage.save();
        }

        UpdatePetData => {
            assert (
                in.senderAddress == storage.collectionAddress ||
                in.senderAddress == storage.breedingContract
            ) throw ERR_NOT_FROM_COLLECTION;

            var pet = msg.petData.load();
            pet.ownerAddress = storage.ownerAddress;
            storage.metadataUri = pet.metadataUri;
            storage.petData = pet.toCell();
            storage.breedCooldownUntil = pet.breedCooldownUntil;
            storage.save();
        }

        NftMint => {
            // Mint is valid only during the uninitialized branch above.
            throw ERR_INVALID_MESSAGE;
        }

        else => {
            if (in.body.isEmpty()) {
                return;
            }
            throw ERR_INVALID_MESSAGE;
        }
    }
}

struct NftDataReply {
    initialized: int
    itemIndex: uint256
    collectionAddress: address
    ownerAddress: address?
    individualContent: cell
}

get fun get_nft_data(): NftDataReply {
    val loadingStorage = startLoadingPetNftStorage();
    if (!loadingStorage.isInitialized()) {
        val deployment = loadingStorage.parseNotInitialized();
        return {
            initialized: 0,
            itemIndex: deployment.itemIndex as uint256,
            collectionAddress: deployment.collectionAddress,
            ownerAddress: null,
            individualContent: emptyCell(),
        };
    }

    val storage = loadingStorage.parseInitialized();
    return {
        initialized: 1,
        itemIndex: storage.itemIndex as uint256,
        collectionAddress: storage.collectionAddress,
        ownerAddress: storage.ownerAddress,
        individualContent: storage.metadataUri,
    };
}

get fun get_pet_data(): PetData {
    val loadingStorage = startLoadingPetNftStorage();
    assert (loadingStorage.isInitialized()) throw ERR_NOT_INITIALIZED;
    val storage = loadingStorage.parseInitialized();
    return storage.petData.load();
}

get fun get_breed_cooldown(): uint64 {
    val loadingStorage = startLoadingPetNftStorage();
    assert (loadingStorage.isInitialized()) throw ERR_NOT_INITIALIZED;
    val storage = loadingStorage.parseInitialized();
    return storage.breedCooldownUntil;
}

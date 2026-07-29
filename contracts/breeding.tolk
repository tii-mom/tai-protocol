// TAI Protocol - Breeding contract
// Language: Tolk 1.0
//
// Incoming message formats:
// 1. transfer_notification#7362d09c query_id:uint64 amount:coins
//    sender:address forward_payload:(remaining bits and refs)
//    The forward payload is BreedRequestPayload:
//    parent1Address:address parent2Address:address.
// 2. request_breed#00000010 query_id:uint64 parent1Address:address
//    parent2Address:address fee:coins initiator:address. This is the
//    explicit form for a call made by the configured TAI wallet.
// 3. pet_data_response#00000006 queryId:uint64 petData:^Cell<PetData>.
//
// Outgoing message formats:
// 1. get_pet_data#00000005 queryId:uint64.
// 2. set_breed_cooldown#00000004 queryId:uint64 cooldownUntil:uint64.
// 3. factory_mint_pet#00000020 queryId:uint64 ownerAddress:address
//    petData:^Cell<PetData>.
// 4. breeding_burn#00000017 queryId:uint64 feeAmount:coins
//    burnAmount:coins responseDestination:address?.

import "pet_nft"

const BREED_OP_REQUEST = 0x00000010
const BREED_OP_FACTORY_MINT = 0x00000020
const BREED_OP_BURN = 0x00000017

const FIXED_BURN_AMOUNT = 10_000_000_000 // 10 TAI, in the 9-decimal unit
const FIXED_BREEDING_FEE = 20_000_000_000 // 20 TAI; 50% is burned
const BREEDING_COOLDOWN_SECONDS = 86_400 // 24 hours
const MIN_ACTION_VALUE = 20_000_000 // 0.02 TON in nanotons
const MAX_QUALITY = 4
const FEE_BURN_PERCENT = 50

const BREED_ERR_INVALID_MESSAGE = 700
const BREED_ERR_NOT_FROM_TAI_WALLET = 701
const BREED_ERR_INVALID_PARENT = 702
const BREED_ERR_NOT_ENOUGH_FEE = 703
const BREED_ERR_IN_PROGRESS = 704
const BREED_ERR_PARENT_RESPONSE = 705
const BREED_ERR_PARENT_OWNER = 706

struct BreedRequestPayload {
    parent1Address: address
    parent2Address: address
}

struct (0x7362d09c) JettonTransferNotification {
    queryId: uint64
    amount: coins
    senderAddress: address
    forwardPayload: RemainingBitsAndRefs
}

struct (0x00000010) RequestBreed {
    queryId: uint64
    parent1Address: address
    parent2Address: address
    fee: coins
    initiator: address
}

struct (0x00000020) FactoryMintPet {
    queryId: uint64
    ownerAddress: address
    petData: Cell<PetData>
}

struct (0x00000017) BreedingBurn {
    queryId: uint64
    feeAmount: coins
    burnAmount: coins
    responseDestination: address?
}

struct BreedParticipants {
    parent1Address: address
    parent2Address: address
    initiator: address
}

struct BreedPending {
    requestId: uint64
    queryId: uint64
    participants: Cell<BreedParticipants>
    fee: coins
    parent1Data: Cell<PetData>?
    parent2Data: Cell<PetData>?
}

struct BreedingConfig {
    taiMasterAddress: address
    taiWalletAddress: address
    petFactoryAddress: address
}

struct BreedingStorage {
    ownerAddress: address
    config: Cell<BreedingConfig>
    requestNonce: uint64
    pending: Cell<BreedPending>?
}

type BreedingIncomingMessage =
    | JettonTransferNotification
    | RequestBreed
    | PetDataResponse

contract Breeding {
    storage: BreedingStorage
    incomingMessages: BreedingIncomingMessage
}

fun BreedingStorage.load() {
    return BreedingStorage.fromCell(contract.getData());
}

fun BreedingStorage.save(self) {
    contract.setData(self.toCell());
}

fun clamp(value: int, lower: int, upper: int): int {
    return max(lower, min(value, upper));
}

// The offset is sampled independently for every dimension. The base is the
// integer mean, while the signed mutation is between 0% and 15% of that mean.
fun breedDimension(parent1: int, parent2: int, lower: int, upper: int): int {
    val mean = (parent1 + parent2) / 2;
    val percent = random.range(16);
    val offset = mean * percent / 100;
    val signedOffset = random.range(2) == 0 ? offset : -offset;
    return clamp(mean + signedOffset, lower, upper);
}

fun generateOffspring(parent1: PetData, parent2: PetData, ownerAddress: address): PetData {
    random.initialize();

    var quality = max(parent1.quality, parent2.quality);
    if (quality < MAX_QUALITY && random.range(100) < 10) {
        quality += 1;
    }

    val species = random.range(2) == 0 ? parent1.species : parent2.species;
    val personality = random.range(2) == 0 ? parent1.personality : parent2.personality;

    return {
        ownerAddress,
        metadataUri: parent1.metadataUri,
        species,
        quality,
        generation: max(parent1.generation, parent2.generation) + 1,
        // Five genetic dimensions: growth rate and four aptitudes.
        growthRate: breedDimension(parent1.growthRate, parent2.growthRate, 100, 130),
        aptAtk: breedDimension(parent1.aptAtk, parent2.aptAtk, 0, 100),
        aptDef: breedDimension(parent1.aptDef, parent2.aptDef, 0, 100),
        aptSpd: breedDimension(parent1.aptSpd, parent2.aptSpd, 0, 100),
        aptInt: breedDimension(parent1.aptInt, parent2.aptInt, 0, 100),
        skillSlots: clamp(max(parent1.skillSlots, parent2.skillSlots), 1, 6),
        personality,
        level: 1,
        exp: 0,
        breedCooldownUntil: 0,
        totalEarned: 0,
    };
}

fun expectedFeeBurn(feeAmount: coins): coins {
    return feeAmount * FEE_BURN_PERCENT / 100;
}

fun sendPetDataRequest(parentAddress: address, queryId: uint64) {
    val request = createMessage({
        bounce: BounceMode.Only256BitsOfBody,
        dest: parentAddress,
        value: MIN_ACTION_VALUE,
        body: GetPetDataRequest { queryId },
    });
    request.send(SEND_MODE_PAY_FEES_SEPARATELY);
}

fun startBreed(
    mutate storage: BreedingStorage,
    queryId: uint64,
    parent1Address: address,
    parent2Address: address,
    initiator: address,
    fee: coins,
) {
    assert (parent1Address != parent2Address) throw BREED_ERR_INVALID_PARENT;
    assert (fee == FIXED_BREEDING_FEE) throw BREED_ERR_NOT_ENOUGH_FEE;
    assert (expectedFeeBurn(fee) == FIXED_BURN_AMOUNT) throw BREED_ERR_NOT_ENOUGH_FEE;
    assert (storage.pending == null) throw BREED_ERR_IN_PROGRESS;

    storage.requestNonce += 1;
    val participants: BreedParticipants = {
        parent1Address,
        parent2Address,
        initiator,
    };
    val pending: BreedPending = {
        requestId: storage.requestNonce,
        queryId,
        participants: participants.toCell(),
        fee,
        parent1Data: null,
        parent2Data: null,
    };
    storage.pending = pending.toCell();
    storage.save();

    sendPetDataRequest(parent1Address, queryId);
    sendPetDataRequest(parent2Address, queryId);
}

fun finishBreed(mutate storage: BreedingStorage, pending: BreedPending) {
    assert (
        pending.parent1Data != null && pending.parent2Data != null
    ) throw BREED_ERR_PARENT_RESPONSE;

    val parent1 = pending.parent1Data!.load();
    val parent2 = pending.parent2Data!.load();
    val participants = pending.participants.load();
    val now = blockchain.now();

    assert (parent1.breedCooldownUntil <= now) throw BREED_ERR_INVALID_PARENT;
    assert (parent2.breedCooldownUntil <= now) throw BREED_ERR_INVALID_PARENT;
    assert (
        participants.initiator == parent1.ownerAddress ||
        participants.initiator == parent2.ownerAddress
    ) throw BREED_ERR_PARENT_OWNER;

    val child = generateOffspring(parent1, parent2, participants.initiator);
    val cooldownUntil = now + BREEDING_COOLDOWN_SECONDS;

    storage.pending = null;
    storage.save();

    val cooldown1 = createMessage({
        bounce: BounceMode.NoBounce,
        dest: participants.parent1Address,
        value: MIN_ACTION_VALUE,
        body: SetBreedCooldown {
            queryId: pending.requestId,
            cooldownUntil,
        },
    });
    cooldown1.send(SEND_MODE_PAY_FEES_SEPARATELY);

    val cooldown2 = createMessage({
        bounce: BounceMode.NoBounce,
        dest: participants.parent2Address,
        value: MIN_ACTION_VALUE,
        body: SetBreedCooldown {
            queryId: pending.requestId,
            cooldownUntil,
        },
    });
    cooldown2.send(SEND_MODE_PAY_FEES_SEPARATELY);

    val mint = createMessage({
        bounce: BounceMode.Only256BitsOfBody,
        dest: storage.config.load().petFactoryAddress,
        value: MIN_ACTION_VALUE,
        body: FactoryMintPet {
            queryId: pending.requestId,
            ownerAddress: participants.initiator,
            petData: child.toCell(),
        },
    });
    mint.send(SEND_MODE_PAY_FEES_SEPARATELY);

    // The TAI master verifies that this is the configured breeding contract,
    // then forwards a standard burn request to this contract's TAI wallet.
    val burn = createMessage({
        bounce: BounceMode.Only256BitsOfBody,
        dest: storage.config.load().taiMasterAddress,
        value: 0,
        body: BreedingBurn {
            queryId: pending.requestId,
            feeAmount: pending.fee,
            burnAmount: FIXED_BURN_AMOUNT,
            responseDestination: contract.getAddress(),
        },
    });
    burn.send(SEND_MODE_CARRY_ALL_REMAINING_MESSAGE_VALUE);
}

fun onInternalMessage(in: InMessage) {
    val msg = lazy BreedingIncomingMessage.fromSlice(in.body);
    var storage = lazy BreedingStorage.load();
    val config = storage.config.load();

    match (msg) {
        JettonTransferNotification => {
            assert (
                in.senderAddress == config.taiWalletAddress
            ) throw BREED_ERR_NOT_FROM_TAI_WALLET;
            assert (msg.forwardPayload.remainingBitsCount() > 0) throw BREED_ERR_INVALID_MESSAGE;

            // The amount in transfer_notification is the Jetton fee. The
            // payload carries only the two parent NFT addresses.
            val payload = BreedRequestPayload.fromSlice(msg.forwardPayload);
            startBreed(
                mutate storage,
                msg.queryId,
                payload.parent1Address,
                payload.parent2Address,
                msg.senderAddress,
                msg.amount,
            );
        }

        RequestBreed => {
            assert (
                in.senderAddress == config.taiWalletAddress
            ) throw BREED_ERR_NOT_FROM_TAI_WALLET;
            startBreed(
                mutate storage,
                msg.queryId,
                msg.parent1Address,
                msg.parent2Address,
                msg.initiator,
                msg.fee,
            );
        }

        PetDataResponse => {
            assert (storage.pending != null) throw BREED_ERR_PARENT_RESPONSE;
            var pending = storage.pending!.load();
            assert (msg.queryId == pending.queryId) throw BREED_ERR_PARENT_RESPONSE;

            val participants = pending.participants.load();
            if (in.senderAddress == participants.parent1Address) {
                assert (pending.parent1Data == null) throw BREED_ERR_PARENT_RESPONSE;
                pending.parent1Data = msg.petData;
            } else if (in.senderAddress == participants.parent2Address) {
                assert (pending.parent2Data == null) throw BREED_ERR_PARENT_RESPONSE;
                pending.parent2Data = msg.petData;
            } else {
                throw BREED_ERR_PARENT_RESPONSE;
            }

            if (pending.parent1Data == null || pending.parent2Data == null) {
                storage.pending = pending.toCell();
                storage.save();
                return;
            }

            finishBreed(mutate storage, pending);
        }

        else => {
            if (in.body.isEmpty()) {
                return;
            }
            throw BREED_ERR_INVALID_MESSAGE;
        }
    }
}

get fun get_breeding_config(): (address, address, address, address, coins) {
    val storage = lazy BreedingStorage.load();
    val config = storage.config.load();
    return (
        config.taiMasterAddress,
        config.taiWalletAddress,
        config.petFactoryAddress,
        contract.getAddress(),
        FIXED_BURN_AMOUNT,
    );
}

get fun get_pending_request(): BreedPending? {
    val storage = lazy BreedingStorage.load();
    if (storage.pending == null) {
        return null;
    }
    return storage.pending!.load();
}

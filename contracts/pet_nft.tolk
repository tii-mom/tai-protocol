// TAI Protocol - Pet NFT Contract
// Standard: TEP-62 (TON NFT)
// Language: Tolk 1.0

// === Storage ===

struct PetData {
    owner: address,
    name: slice,
    species: int,          // 0=dragon, 1=falcon, 2=wolf, 3=cat, 4=rabbit, 5=turtle
    quality: int,          // 0=N, 1=R, 2=SR, 3=SSR, 4=UR
    generation: int,       // 0=genesis
    growthRate: int,       // 100-130
    aptAtk: int,           // 0-100
    aptDef: int,           // 0-100
    aptSpd: int,           // 0-100
    aptInt: int,           // 0-100
    skillSlots: int,       // 1-6
    personality: int,      // enum
    level: int,
    exp: int,
    breedCooldownUntil: int,
    totalEarned: int,
}

// === Constants ===

const OP_MINT: int = 0x01;
const OP_TRANSFER: int = 0x02;
const OP_UPDATE_ATTRS: int = 0x03;
const OP_SET_BREED_COOLDOWN: int = 0x04;
const OP_GET_DATA: int = 0x05;

// === Entrypoint ===

fun onInternalMessage(myBalance: int, msgValue: int, inMsgFull: cell, inMsgBody: slice) {
    // TODO: Parse op code and dispatch
    // TODO: Implement mint (only from collection contract)
    // TODO: Implement transfer (only from owner)
    // TODO: Implement update_attrs (only from authorized contracts)
    // TODO: Implement get_data (read-only query)
}

fun getMethods() {
    // get_nft_data() -> TEP-62 standard
    // get_pet_data() -> PetData full struct
}

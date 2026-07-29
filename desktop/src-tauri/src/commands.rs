use serde::Serialize;

#[derive(Serialize)]
pub struct PetState {
    pub id: String,
    pub name: String,
    pub level: u32,
    pub energy: f32,
    pub mood: f32,
    pub status: String,
    pub animation: String,
}

#[derive(Serialize)]
pub struct EarningsSummary {
    pub today_tai: f64,
    pub today_usdt: f64,
    pub total_tai: f64,
    pub active_bounties: u32,
}

/// 获取当前宠物状态（从后端 API 拉取）
#[tauri::command]
pub async fn get_pet_state() -> Result<PetState, String> {
    // TODO: 调用后端 /api/v1/pet/mine 获取活跃宠物
    Ok(PetState {
        id: "gen0-001".to_string(),
        name: "Mecha-Alpha".to_string(),
        level: 1,
        energy: 100.0,
        mood: 85.0,
        status: "idle".to_string(),
        animation: "idle_breathing".to_string(),
    })
}

/// 向宠物发送指令（喂食、休息、执行任务等）
#[tauri::command]
pub async fn send_pet_command(command: String) -> Result<String, String> {
    // TODO: 转发到后端或直接触发 Agent 行为
    Ok(format!("Command '{}' acknowledged", command))
}

/// 获取收益摘要
#[tauri::command]
pub async fn get_earnings_summary() -> Result<EarningsSummary, String> {
    // TODO: 调用后端 /api/v1/user/balance
    Ok(EarningsSummary {
        today_tai: 0.0,
        today_usdt: 0.0,
        total_tai: 0.0,
        active_bounties: 0,
    })
}

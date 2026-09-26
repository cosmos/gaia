// Model of the v29.4.0 refund: community pool -> ch_user, locked in a
// DelayedVestingAccount until UNLOCK. Numbers are scaled down so the
// simulator can explore them. See upgrades.go for the real thing.
//
// quint run app/upgrades/v29_4_0/refund.qnt --invariant=safety --max-samples=20000
module refund {
  pure val REFUND = 5
  pure val UNLOCK = 3
  pure val MAX_TIME = 5

  type AccKind = NoAcc | BaseAcc | VestingAcc | ModuleAcc
  type Phase = Pre | Done | Halted

  pure def max(a: int, b: int): int = if (a > b) a else b
  pure def min(a: int, b: int): int = if (a < b) a else b

  // chain
  var now: int
  var phase: Phase
  // community pool accounting, and the distribution module coins behind it
  var pool: int
  var distrBal: int
  // ch_user
  var kind: AccKind
  var blocked: bool
  var bal: int
  var origVesting: int
  var endTime: int
  var delVesting: int
  var delFree: int
  var bonded: int
  // coins ch_user sent to anyone else
  var sentOut: int

  // snapshot of the pre upgrade state
  var init0: { pool: int, distrBal: int, kind: AccKind, bal: int }

  // SDK vesting math, mirrors x/auth/vesting BaseVestingAccount.
  def vestingCoins = if (kind == VestingAcc and now < endTime) origVesting else 0
  def locked = max(0, vestingCoins - delVesting)
  def spendable = bal - locked

  // Same pre conditions as refundFromCommunityPool.
  def canUpgrade =
    now < UNLOCK and
    not(blocked) and
    kind == BaseAcc and
    pool >= REFUND and
    distrBal >= REFUND

  action init = {
    nondet p = 0.to(8).oneOf()
    nondet d = 0.to(8).oneOf()
    nondet k = Set(NoAcc, BaseAcc, VestingAcc, ModuleAcc).oneOf()
    nondet e = 0.to(3).oneOf()
    nondet t = 0.to(MAX_TIME).oneOf()
    nondet b = Set(true, false).oneOf()
    val e2 = if (k == NoAcc) 0 else e
    all {
      now' = t,
      phase' = Pre,
      pool' = p,
      distrBal' = d,
      kind' = k,
      blocked' = b,
      bal' = e2,
      origVesting' = 0,
      endTime' = 0,
      delVesting' = 0,
      delFree' = 0,
      bonded' = 0,
      sentOut' = 0,
      init0' = { pool: p, distrBal: d, kind: k, bal: e2 },
    }
  }

  action unchangedAcct = all {
    blocked' = blocked, origVesting' = origVesting, endTime' = endTime,
    delVesting' = delVesting, delFree' = delFree, bonded' = bonded,
    sentOut' = sentOut, init0' = init0,
  }

  action upgrade = all {
    phase == Pre,
    now' = now,
    if (canUpgrade) all {
      phase' = Done,
      pool' = pool - REFUND,
      distrBal' = distrBal - REFUND,
      kind' = VestingAcc,
      bal' = bal + REFUND,
      origVesting' = REFUND,
      endTime' = UNLOCK,
      blocked' = blocked, delVesting' = delVesting, delFree' = delFree,
      bonded' = bonded, sentOut' = sentOut, init0' = init0,
    } else all {
      // Handler errors, chain halts, no state written.
      phase' = Halted,
      pool' = pool, distrBal' = distrBal, kind' = kind, bal' = bal,
      unchangedAcct,
    },
  }

  action tick = all {
    now < MAX_TIME,
    now' = now + 1,
    phase' = phase, pool' = pool, distrBal' = distrBal, kind' = kind, bal' = bal,
    unchangedAcct,
  }

  action send(a: int): bool = all {
    phase == Done,
    a <= spendable,
    bal' = bal - a,
    sentOut' = sentOut + a,
    now' = now, phase' = phase, pool' = pool, distrBal' = distrBal, kind' = kind,
    blocked' = blocked, origVesting' = origVesting, endTime' = endTime,
    delVesting' = delVesting, delFree' = delFree, bonded' = bonded, init0' = init0,
  }

  // x/bank DelegateCoins checks the full balance, then TrackDelegation
  // splits it into vesting and free parts.
  action delegate(a: int): bool = {
    val dv = min(max(0, vestingCoins - delVesting), a)
    all {
      phase == Done,
      a <= bal,
      bal' = bal - a,
      bonded' = bonded + a,
      delVesting' = delVesting + dv,
      delFree' = delFree + (a - dv),
      now' = now, phase' = phase, pool' = pool, distrBal' = distrBal, kind' = kind,
      blocked' = blocked, origVesting' = origVesting, endTime' = endTime,
      sentOut' = sentOut, init0' = init0,
    }
  }

  // TrackUndelegation takes from free first, then vesting.
  action undelegate(a: int): bool = {
    val x = min(delFree, a)
    val y = min(delVesting, a - x)
    all {
      phase == Done,
      a <= bonded,
      bal' = bal + a,
      bonded' = bonded - a,
      delFree' = delFree - x,
      delVesting' = delVesting - y,
      now' = now, phase' = phase, pool' = pool, distrBal' = distrBal, kind' = kind,
      blocked' = blocked, origVesting' = origVesting, endTime' = endTime,
      sentOut' = sentOut, init0' = init0,
    }
  }

  action step = {
    nondet a = 1.to(REFUND + 3).oneOf()
    any { upgrade, tick, send(a), delegate(a), undelegate(a) }
  }

  // --- properties ---

  // Coins are only moved, never made or lost.
  val conservation =
    bal + bonded + sentOut + distrBal == init0.bal + init0.distrBal

  // Upgrade only goes through when every precondition held.
  val onlyWhenSafe = phase == Done implies and {
    init0.kind == BaseAcc,
    not(blocked),
    init0.pool >= REFUND,
    init0.distrBal >= REFUND,
  }

  // Pool is debited exactly REFUND, and the account is set up right.
  val exactMove = phase == Done implies and {
    pool == init0.pool - REFUND,
    kind == VestingAcc,
    origVesting == REFUND,
    endTime == UNLOCK,
  }

  // A failed upgrade writes nothing.
  val haltIsNoop = phase == Halted implies and {
    pool == init0.pool,
    distrBal == init0.distrBal,
    kind == init0.kind,
    bal == init0.bal,
  }

  // Staking only: before UNLOCK ch_user can't send out more than they had pre upgrade.
  val refundLocked = (phase == Done and now < UNLOCK) implies sentOut <= init0.bal

  // After UNLOCK everything is liquid.
  val unlockedAfter = (phase == Done and now >= UNLOCK) implies spendable == bal

  val safety = and {
    conservation,
    onlyWhenSafe,
    exactMove,
    haltIsNoop,
    refundLocked,
    unlockedAfter,
  }

  // Witnesses. These should be violated, showing the states are reachable.
  val noDone = phase != Done
  val noHalt = phase != Halted
  val noStakeLocked = not(phase == Done and now < UNLOCK and delVesting > 0)
}

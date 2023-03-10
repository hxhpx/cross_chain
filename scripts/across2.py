from common import *


FundsDeposited = "0x4a4fc49abd237bfd7f4ac82d6c7a284c69daaea5154430cff04ad7482c6c4254"
FilledRelay = "0x56450a30040c51955338a4a9fbafcf94f7ca4b75f4cd83c2f5e29ef77fbe0a3a"

sigs = {
    FilledRelay: "(uint256,uint256,uint256,uint256,uint256,uint256,uint64,uint64,uint64,uint32,address,address,bool)",
    FundsDeposited: "(uint256,uint256,uint256,uint64,uint32,address)"
}

AcrossContracts = {
    "ethereum": [
        #"0x4d9079bb4165aeb4084c526a32695dcfd2f77381", "0x931a43528779034ac9eb77df799d133557406176",
    ],
    "polygon": [
		#"0x69b5c72837769ef1e7c164abc6515dcff217f920", "0xd3ddacae5afb00f9b9cd36ef0ed7115d7f0b584c",
    ],
    "arbitrum": [
		#"0xb88690461ddbab6f04dfad7df66b7725942feb9c", "0xe1c367e2b576ac421a9f46c9cc624935730c36aa",
    ],
    "optimism": [
		#"0x59485d57eecc4058f7831f46ee83a7078276b4ae", "0xa420b2d1c0841415a695b81e5b867bcd07dff8c9",
    ],
    "boba":[
        #"0xBbc6009fEfFc27ce705322832Cb2068F8C1e0A58"
    ],
}

Web3QueryStartBlock = {
    "ethereum": 13719989,
    "bsc": 13099216,
    "polygon": 22006535,
    "fantom": 23658021,
    "arbitrum": 3483123,
    "avalanche": 7660096,
    "optimism": 646444,
    "boba": 0,
}

Web3QueryBatchSize = {
    # alchemy
    "ethereum": 10000,
    "polygon": 10000,
    "optimism": 10000,
    "arbitrum": 10000,

    # getblock
    "bsc": 1000,
    "fantom": 10000,
    "avalanche": 10000,

    "moonbeam":5000,
    "moonriver":5000,
    "cronos":5000,
    "boba":5000,
    "klaytn":5000,
    "dogechain":5000,
    "harmony":5000,
    "dfk":5000,
    "metis":5000,
    "canto":5000,
}

tableName = "across"
Base = declarative_base()


class AcrossSchema(Base):
    __tablename__ = tableName

    id = Column("id", Integer, primary_key=True)
    chain = Column("chain", String)
    blockNumber = Column("block_number", Integer)
    txIndex = Column("tx_index", Integer)
    txHash = Column("hash", String)
    actionId = Column("action_id", Integer)
    direction = Column("direction", String)
    contract = Column("contract", String)
    fromChain = Column("from_chain", String)
    token = Column("token", String)
    fromAddress = Column("from_address", String)
    amount = Column("amount", Numeric)
    totalAmount = Column("total_amount", Numeric)
    fillAmount = Column("fill_amount", Numeric)
    toAddress = Column("to_address", String)
    toChain = Column("to_chain", String)
    matchTag = Column("match_tag", String)
    detail = Column("detail", JSON)



def _query_and_parse_from_web3(chain, start_block, end_block):
    topics = [[FilledRelay, FundsDeposited]]
    addresses = AcrossContracts[chain]
    items = query_web3_auto(chain, start_block, end_block, topics, addresses)
    objs = []

    for item in items:
        topics = item["topics"]
        topic0 = topics[0]
        data = item["data"]
        direction = ""
        toChainID = -1
        fromChainID = -1
        matchTag = ""
        relayer = ""
        repayChainId = -1
        totalAmount = -1
        fillAmount = -1

        # ??????direction??????token??????
        if topic0 == FundsDeposited:
            (amount, fromChainID, toChainID, relayerFeePct, quoteTimeStamp,
             _to) = decode_single(sigs[topic0], bytes.fromhex(data[2:]))
            matchTag = int(topics[1], 16)
            token = "0x" + (topics[2])[26:]
            _from = "0x" + (topics[3])[26:]
            direction = "out"

        elif topic0 == FilledRelay:
            (amount, totalAmount, fillAmount, repayChainId, fromChainID, toChainID, relayerFeePct,
             appliedRelayerFeePct, realizedLpFeePct, matchTag, token, _to,
             isSlowRelay) = decode_single(sigs[topic0], bytes.fromhex(data[2:]))
            relayer = "0x" + (topics[1])[26:]
            _from = "0x" + (topics[2])[26:]
            direction = "in"
        else:
            continue

        detail = {
            "relayerFeePct": relayerFeePct,
            "relayer": relayer,
            "repayChainId": repayChainId,
        }


        objs.append(
            AcrossSchema(chain=chain,
                          blockNumber=item["blockNumber"],
                          txIndex=item["transactionIndex"],
                          txHash=item["transactionHash"],
                          actionId=item["logIndex"],
                          direction=direction,
                          contract=item["address"],
                          fromChain=fromChainID,
                          token=token,
                          fromAddress=_from,
                          amount=amount,
                          totalAmount=totalAmount,
                          fillAmount=fillAmount,
                          toAddress=_to,
                          toChain=toChainID,
                          matchTag=matchTag,
                          detail=detail,
                          ))
    return objs


def acrossMain():
    if not inspect(engine).has_table(tableName):
        print(f"table {tableName} not exists")
        Base.metadata.create_all(engine, checkfirst=True)

    chains = ["boba", "arbitrum"]

    for chain in chains:
        session = Session()
        stmt = select(AcrossSchema.blockNumber).where(AcrossSchema.chain == chain).order_by(
            AcrossSchema.blockNumber.desc()).limit(1)
        latest_block_in_db = session.scalar(stmt)
        print(f"latest block in database for {chain}: {latest_block_in_db}")

        latest_block_from_web3 = query_web3_latest_block(chain)
        print(f"latest block from web3 for {chain}: {latest_block_from_web3}")

        if latest_block_in_db is not None:
            start_block = latest_block_in_db
        elif chain in Web3QueryStartBlock:
            start_block = Web3QueryStartBlock[chain]
        else:
            start_block = 1

        end_block = latest_block_from_web3
        step = Web3QueryBatchSize[chain]
        inserted = 0
        pbar = tqdm(range(start_block, end_block, step))
        for i in pbar:
            objs = _query_and_parse_from_web3(chain, i + 1, i + step)
            session.bulk_save_objects(objs)
            session.commit()
            inserted += len(objs)
            pbar.set_postfix({"last_inserted": len(objs)})
        print(f"objects inserted: {inserted}")


if __name__ == "__main__":
    acrossMain()

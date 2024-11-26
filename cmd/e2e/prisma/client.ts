import { PrismaClient } from "@prisma/client"
// Database connection will increase once we create more BDD scenarios.
// Per Prisma best practice: create a single prisma instance for each worker to perform database operations
// context: https://www.prisma.io/docs/orm/prisma-client/setup-and-configuration/databases-connections#re-using-a-single-prismaclient-instance
let prisma = new PrismaClient()
export default prisma
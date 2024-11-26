import prisma from "./client"

export class dbOPS {
    // delete test data in dev environment only
    async deleteUsers() {
        const baseURL = `${process.env.BASEURL}`.toLowerCase()
        if (((`${process.env.ENV}` != "production" || `${process.env.ENV}` != "staging")) && (baseURL.toLowerCase().includes('localhost') || baseURL.toLowerCase().includes("127.0.0.1"))) {
            const deleteUsers = prisma.users.deleteMany()
            const deleteUserRoles = prisma.users_roles.deleteMany()
            await prisma.$transaction([deleteUserRoles, deleteUsers])
        } else {
            console.log(`Skipping deletion test data, baseURL: ${process.env.BASEURL} does not target local environment`)
        }
    }
}

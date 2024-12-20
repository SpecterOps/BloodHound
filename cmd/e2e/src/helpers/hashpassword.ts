import { hash, argon2id } from 'argon2';

/**
 * @param password a plain-text password
 *
 * Returns a string suitable for use as the 'digest' of an auth_secrets record in the postgres db
 * Matches the same algorithm used by the BH API when creating or updating auth_secrets
 */
const hashPassword = async (password: string) => {
    const hashedPassword = await hash(password, {
        type: argon2id,
        timeCost: 3,
        memoryCost: 1048576,
        parallelism: 32,
        hashLength: 16
    });
    const hashWithPadding = hashedPassword
        .split('$')
        .map((substr, i) => {
            if (i === 4 || i === 5) return substr.padEnd(24, '=');
            return substr;
        })
        .join('$');
    return hashWithPadding;
};

export default hashPassword;

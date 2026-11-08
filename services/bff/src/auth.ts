import type { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';

const SECRET = process.env.JWT_SECRET ?? '';

export type AuthUser = { id: string; username: string };

// Lê o token  
function readUser(req: Request): AuthUser | null {
    const header = req.headers.authorization ?? '';
    const [kind, token] = header.split(' ');
    if (kind !== 'Bearer' || !token) return null;
    try {
        const p = jwt.verify(token, SECRET) as { user_id: string | number; username: string };
        return { id: String(p.user_id), username: p.username };
    } catch {
        return null;
    }
}

// Rotas que exigem login
export function requireAuth(req: Request, res: Response, next: NextFunction) {
    const user = readUser(req);
    if (!user) return res.status(401).json({ error: { code: 'UNAUTHORIZED', message: 'Token ausente ou inválido' } });
    res.locals.user = user;
    next();
}

// Rotas onde o login é opcional  
export function optionalAuth(req: Request, res: Response, next: NextFunction) {
    res.locals.user = readUser(req);
    next();
}
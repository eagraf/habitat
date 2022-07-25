import * as notesService from "./notes.service";
import { Request, Response } from "express";
import OrbitDB from "orbit-db";

export async function get(req: Request, res: Response) {
    const notes = notesService.getNotes()
    res.send(Array.from(notes.keys()))
}

export async function create(req: Request, res: Response) {
    const name = req.query.name as string
    const addr = await notesService.addNote(name)
    res.send(addr)
    
}

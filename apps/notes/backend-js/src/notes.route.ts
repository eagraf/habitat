import express from 'express'
import * as notesController from './notes.controller'

const router = express.Router()

router.get('/', notesController.get)

router.post('/', notesController.create)

export default router